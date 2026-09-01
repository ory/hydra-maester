// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/hydra"
)

const (
	DefaultClientID  = "CLIENT_ID"
	DefaultSecretKey = "CLIENT_SECRET"
	FinalizerName    = "finalizer.ory.hydra.sh"

	DefaultNamespace = "default"

	// SecretWaitBaseInterval is the delay before the first re-check of a secret that
	// does not exist yet. Every subsequent re-check doubles the delay, up to
	// SecretWaitMaxInterval.
	SecretWaitBaseInterval = 15 * time.Second
	SecretWaitMaxInterval  = 5 * time.Minute
)

var (
	ClientIDKey     = DefaultClientID
	ClientSecretKey = DefaultSecretKey
)

type clientKey struct {
	url            string
	port           int
	endpoint       string
	forwardedProto string
}

// OAuth2ClientFactory is a function that creates oauth2 client.
// The OAuth2ClientReconciler defaults to use hydra.New and the factory allows
// to override this behavior for mocks during tests.
type OAuth2ClientFactory func(
	spec hydrav1alpha1.OAuth2ClientSpec,
	tlsTrustStore string,
	insecureSkipVerify bool,
) (hydra.Client, error)

// OAuth2ClientReconciler reconciles a OAuth2Client object.
type OAuth2ClientReconciler struct {
	client.Client
	HydraClient         hydra.Client
	Log                 logr.Logger
	ControllerNamespace string

	// RequireExistingSecret is the controller-wide default for
	// OAuth2ClientSpec.RequireExistingSecret.
	RequireExistingSecret bool

	oauth2Clients       map[clientKey]hydra.Client
	oauth2ClientFactory OAuth2ClientFactory
	mu                  sync.Mutex

	// secretWaitAttempts counts, per object, how often we requeued because the
	// referenced secret did not exist yet. It is only used to compute the backoff.
	secretWaitAttempts map[types.NamespacedName]int
	secretWaitMu       sync.Mutex
}

// Options represent options to pass to the oauth2 client reconciler.
type Options struct {
	Namespace             string
	OAuth2ClientFactory   OAuth2ClientFactory
	RequireExistingSecret bool
}

// Option is a functional option.
type Option func(*Options)

func init() {
	if os.Getenv("CLIENT_ID_KEY") != "" {
		ClientIDKey = os.Getenv("CLIENT_ID_KEY")
	}
	if os.Getenv("CLIENT_SECRET_KEY") != "" {
		ClientSecretKey = os.Getenv("CLIENT_SECRET_KEY")
	}
}

// WithNamespace sets the kubernetes namespace for the controller.
// The default is "default".
func WithNamespace(ns string) Option {
	return func(o *Options) {
		o.Namespace = ns
	}
}

// WithClientFactory sets a function to create new oauth2 clients during the reconciliation logic.
func WithClientFactory(factory OAuth2ClientFactory) Option {
	return func(o *Options) {
		o.OAuth2ClientFactory = factory
	}
}

// WithRequireExistingSecret sets the controller-wide default for whether the
// controller waits for the secret referenced by an OAuth2Client instead of
// registering the client with a generated secret and creating the secret itself.
// Individual resources can override it via `spec.requireExistingSecret`.
func WithRequireExistingSecret(require bool) Option {
	return func(o *Options) {
		o.RequireExistingSecret = require
	}
}

// New returns a new Oauth2ClientReconciler.
func New(c client.Client, hydraClient hydra.Client, log logr.Logger, opts ...Option) *OAuth2ClientReconciler {
	options := &Options{
		Namespace:           DefaultNamespace,
		OAuth2ClientFactory: hydra.New,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &OAuth2ClientReconciler{
		Client:                c,
		HydraClient:           hydraClient,
		Log:                   log,
		ControllerNamespace:   options.Namespace,
		RequireExistingSecret: options.RequireExistingSecret,
		oauth2Clients:         make(map[clientKey]hydra.Client, 0),
		oauth2ClientFactory:   options.OAuth2ClientFactory,
		secretWaitAttempts:    make(map[types.NamespacedName]int),
	}
}

// +kubebuilder:rbac:groups=hydra.ory.sh,resources=oauth2clients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hydra.ory.sh,resources=oauth2clients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *OAuth2ClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("oauth2client", req.NamespacedName)

	var oauth2client hydrav1alpha1.OAuth2Client
	if err := r.Get(ctx, req.NamespacedName, &oauth2client); err != nil {
		if apierrs.IsNotFound(err) {
			r.resetSecretWait(req.NamespacedName)
			if registerErr := r.unregisterOAuth2Clients(ctx, &oauth2client); registerErr != nil {
				return ctrl.Result{}, registerErr
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check request namespace
	if r.ControllerNamespace != "" {
		r.Log.Info(fmt.Sprintf("ControllerNamespace is set to: %s, working only on items in this namespace. Other namespaces are ignored.", r.ControllerNamespace))
		if req.NamespacedName.Namespace != r.ControllerNamespace {
			r.Log.Info(fmt.Sprintf("Requested resource %s is not in namespace: %s and will be ignored", req.String(), r.ControllerNamespace))
			return ctrl.Result{}, nil
		}
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if oauth2client.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(oauth2client.ObjectMeta.Finalizers, FinalizerName) {
			typeMeta := oauth2client.TypeMeta
			oauth2client.ObjectMeta.Finalizers = append(oauth2client.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &oauth2client); err != nil {
				return ctrl.Result{}, err
			}
			// restore the TypeMeta object as it is removed during Update, but need to be accessed later
			oauth2client.TypeMeta = typeMeta
		}
	} else {
		// The object is being deleted
		r.resetSecretWait(req.NamespacedName)
		if containsString(oauth2client.ObjectMeta.Finalizers, FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.unregisterOAuth2Clients(ctx, &oauth2client); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			oauth2client.ObjectMeta.Finalizers = removeString(oauth2client.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &oauth2client); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil

	}

	var secret apiv1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: oauth2client.Spec.SecretName, Namespace: req.Namespace}, &secret); err != nil {
		if apierrs.IsNotFound(err) {
			if r.requireExistingSecret(&oauth2client) {
				// The secret is managed outside of this controller and has not been
				// created yet. Wait for it instead of registering the client with a
				// generated secret and creating the secret ourselves.
				return r.waitForSecret(ctx, &oauth2client, req.NamespacedName)
			}
			if registerErr := r.registerOAuth2Client(ctx, &oauth2client, nil); registerErr != nil {
				return ctrl.Result{}, registerErr
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	r.resetSecretWait(req.NamespacedName)

	credentials, err := parseSecret(secret, oauth2client.Spec.TokenEndpointAuthMethod)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("secret %s/%s is invalid", secret.Name, secret.Namespace))
		if updateErr := r.updateReconciliationStatusError(ctx, &oauth2client, hydrav1alpha1.StatusInvalidSecret, err); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, nil
	}

	hydraClient, err := r.getHydraClientForClient(oauth2client)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf(
			"hydra address %s:%d%s is invalid",
			oauth2client.Spec.HydraAdmin.URL,
			oauth2client.Spec.HydraAdmin.Port,
			oauth2client.Spec.HydraAdmin.Endpoint,
		))
		if updateErr := r.updateReconciliationStatusError(ctx, &oauth2client, hydrav1alpha1.StatusInvalidHydraAddress, err); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, nil
	}

	fetched, found, err := hydraClient.GetOAuth2Client(string(credentials.ID))
	if err != nil {
		return ctrl.Result{}, err
	}

	if found {
		//conclude reconciliation if the client exists and has not been updated
		if oauth2client.Generation == oauth2client.Status.ObservedGeneration {
			return ctrl.Result{}, nil
		}

		if fetched.Owner != fmt.Sprintf("%s/%s", oauth2client.Name, oauth2client.Namespace) {
			conflictErr := fmt.Errorf("ID provided in secret %s/%s is assigned to another resource", secret.Name, secret.Namespace)
			if updateErr := r.updateReconciliationStatusError(ctx, &oauth2client, hydrav1alpha1.StatusInvalidSecret, conflictErr); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{}, nil
		}

		if updateErr := r.updateRegisteredOAuth2Client(ctx, &oauth2client, credentials); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, nil
	}

	if registerErr := r.registerOAuth2Client(ctx, &oauth2client, credentials); registerErr != nil {
		return ctrl.Result{}, registerErr
	}

	return ctrl.Result{}, nil
}

func (r *OAuth2ClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hydrav1alpha1.OAuth2Client{}).
		Complete(r)
}

// requireExistingSecret reports whether the referenced secret has to exist before
// the client is registered in hydra. The value on the resource takes precedence
// over the controller-wide default.
func (r *OAuth2ClientReconciler) requireExistingSecret(c *hydrav1alpha1.OAuth2Client) bool {
	if c.Spec.RequireExistingSecret != nil {
		return *c.Spec.RequireExistingSecret
	}
	return r.RequireExistingSecret
}

// waitForSecret marks the client as not ready and requeues it with an exponential
// backoff, so that reconciliation resumes once the referenced secret shows up.
func (r *OAuth2ClientReconciler) waitForSecret(ctx context.Context, c *hydrav1alpha1.OAuth2Client, key types.NamespacedName) (ctrl.Result, error) {
	requeueAfter := r.nextSecretWaitInterval(key)

	r.Log.Info(
		fmt.Sprintf("secret %s/%s does not exist yet, waiting for it before registering the client", c.Namespace, c.Spec.SecretName),
		"oauth2client", fmt.Sprintf("%s/%s", c.Namespace, c.Name),
		"requeueAfter", requeueAfter.String(),
	)

	err := fmt.Errorf("secret %s/%s does not exist", c.Namespace, c.Spec.SecretName)
	if updateErr := r.updateReconciliationStatusPending(ctx, c, hydrav1alpha1.StatusSecretNotFound, err); updateErr != nil {
		return ctrl.Result{}, updateErr
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// nextSecretWaitInterval returns the backoff for the next re-check of a missing
// secret, doubling on every consecutive attempt up to SecretWaitMaxInterval.
func (r *OAuth2ClientReconciler) nextSecretWaitInterval(key types.NamespacedName) time.Duration {
	r.secretWaitMu.Lock()
	defer r.secretWaitMu.Unlock()

	if r.secretWaitAttempts == nil {
		r.secretWaitAttempts = make(map[types.NamespacedName]int)
	}

	attempt := r.secretWaitAttempts[key]
	r.secretWaitAttempts[key] = attempt + 1

	interval := SecretWaitBaseInterval
	for i := 0; i < attempt; i++ {
		interval *= 2
		if interval >= SecretWaitMaxInterval {
			return SecretWaitMaxInterval
		}
	}

	return interval
}

// resetSecretWait drops the backoff state for a client that is no longer waiting
// for its secret.
func (r *OAuth2ClientReconciler) resetSecretWait(key types.NamespacedName) {
	r.secretWaitMu.Lock()
	defer r.secretWaitMu.Unlock()

	delete(r.secretWaitAttempts, key)
}

func (r *OAuth2ClientReconciler) registerOAuth2Client(ctx context.Context, c *hydrav1alpha1.OAuth2Client, credentials *hydra.Oauth2ClientCredentials) error {
	if err := r.unregisterOAuth2Clients(ctx, c); err != nil {
		return err
	}

	hydraClient, err := r.getHydraClientForClient(*c)
	if err != nil {
		return err
	}

	oauth2client, err := hydra.FromOAuth2Client(c)
	if err != nil {
		if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusRegistrationFailed, err); updateErr != nil {
			return updateErr
		}

		return fmt.Errorf("failed to construct hydra client for object: %w", err)
	}

	if credentials != nil {
		if _, err := hydraClient.PostOAuth2Client(oauth2client.WithCredentials(credentials)); err != nil {
			if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusRegistrationFailed, err); updateErr != nil {
				return updateErr
			}
		}
		return r.ensureEmptyStatusError(ctx, c)
	}

	created, err := hydraClient.PostOAuth2Client(oauth2client)
	if err != nil {
		if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusRegistrationFailed, err); updateErr != nil {
			return updateErr
		}
		return nil
	}

	clientSecret := apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Spec.SecretName,
			Namespace: c.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: c.TypeMeta.APIVersion,
				Kind:       c.TypeMeta.Kind,
				Name:       c.ObjectMeta.Name,
				UID:        c.ObjectMeta.UID,
			}},
		},
		Data: map[string][]byte{
			ClientIDKey: []byte(*created.ClientID),
		},
	}

	if created.Secret != nil {
		clientSecret.Data[ClientSecretKey] = []byte(*created.Secret)
	}

	if err := r.Create(ctx, &clientSecret); err != nil {
		if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusCreateSecretFailed, err); updateErr != nil {
			return updateErr
		}
	}

	return r.ensureEmptyStatusError(ctx, c)
}

func (r *OAuth2ClientReconciler) updateRegisteredOAuth2Client(ctx context.Context, c *hydrav1alpha1.OAuth2Client, credentials *hydra.Oauth2ClientCredentials) error {
	hydraClient, err := r.getHydraClientForClient(*c)
	if err != nil {
		return err
	}

	oauth2client, err := hydra.FromOAuth2Client(c)
	if err != nil {
		if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusUpdateFailed, err); updateErr != nil {
			return updateErr
		}

		return fmt.Errorf("failed to construct hydra client for object: %w", err)
	}

	if _, err := hydraClient.PutOAuth2Client(oauth2client.WithCredentials(credentials)); err != nil {
		if updateErr := r.updateReconciliationStatusError(ctx, c, hydrav1alpha1.StatusUpdateFailed, err); updateErr != nil {
			return updateErr
		}
	}
	return r.ensureEmptyStatusError(ctx, c)
}

func (r *OAuth2ClientReconciler) unregisterOAuth2Clients(ctx context.Context, c *hydrav1alpha1.OAuth2Client) error {
	// if a required field is empty, that means this is deleted after
	// the finalizers have done their job, so just return
	if (c.Spec.Scope == "" && len(c.Spec.ScopeArray) == 0) || c.Spec.SecretName == "" {
		return nil
	}

	h, err := r.getHydraClientForClient(*c)
	if err != nil {
		return err
	}

	clients, err := h.ListOAuth2Client()
	if err != nil {
		return err
	}

	for _, cJSON := range clients {
		if cJSON.Owner == fmt.Sprintf("%s/%s", c.Name, c.Namespace) {
			if c.Spec.DeletionPolicy == hydrav1alpha1.OAuth2ClientDeletionPolicyOrphan {
				// Do not delete the OAuth2 client.
				r.Log.Info("oauth2 client deletion, leave the row orphan")
				return nil
			}
			if err := h.DeleteOAuth2Client(*cJSON.ClientID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *OAuth2ClientReconciler) updateReconciliationStatusError(ctx context.Context, c *hydrav1alpha1.OAuth2Client, code hydrav1alpha1.StatusCode, err error) error {
	r.Log.Error(err, fmt.Sprintf("error processing client %s/%s ", c.Name, c.Namespace), "oauth2client", "register")

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, c, func() error {
		c.Status.ObservedGeneration = c.Generation
		c.Status.ReconciliationError = hydrav1alpha1.ReconciliationError{
			Code:        code,
			Description: err.Error(),
		}
		c.Status.Conditions = []hydrav1alpha1.OAuth2ClientCondition{
			{
				Type:   hydrav1alpha1.OAuth2ClientConditionReady,
				Status: hydrav1alpha1.ConditionFalse,
			},
		}

		return nil
	})
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("status update failed for client %s/%s ", c.Name, c.Namespace), "oauth2client", "update status")
	}

	return err
}

// updateReconciliationStatusPending reports that reconciliation has not finished
// yet and will be retried. Unlike updateReconciliationStatusError it leaves
// ObservedGeneration untouched, so the current generation is still reconciled
// from scratch once the blocker is gone.
func (r *OAuth2ClientReconciler) updateReconciliationStatusPending(ctx context.Context, c *hydrav1alpha1.OAuth2Client, code hydrav1alpha1.StatusCode, err error) error {
	_, updateErr := controllerutil.CreateOrPatch(ctx, r.Client, c, func() error {
		c.Status.ReconciliationError = hydrav1alpha1.ReconciliationError{
			Code:        code,
			Description: err.Error(),
		}
		c.Status.Conditions = []hydrav1alpha1.OAuth2ClientCondition{
			{
				Type:   hydrav1alpha1.OAuth2ClientConditionReady,
				Status: hydrav1alpha1.ConditionFalse,
			},
		}

		return nil
	})
	if updateErr != nil {
		r.Log.Error(updateErr, fmt.Sprintf("status update failed for client %s/%s ", c.Name, c.Namespace), "oauth2client", "update status")
	}

	return updateErr
}

func (r *OAuth2ClientReconciler) ensureEmptyStatusError(ctx context.Context, c *hydrav1alpha1.OAuth2Client) error {
	_, err := controllerutil.CreateOrPatch(ctx, r.Client, c, func() error {
		c.Status.ObservedGeneration = c.Generation
		c.Status.ReconciliationError = hydrav1alpha1.ReconciliationError{}
		c.Status.Conditions = []hydrav1alpha1.OAuth2ClientCondition{
			{
				Type:   hydrav1alpha1.OAuth2ClientConditionReady,
				Status: hydrav1alpha1.ConditionTrue,
			},
		}

		return nil
	})
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("status update failed for client %s/%s ", c.Name, c.Namespace), "oauth2client", "update status")
	}

	return err
}

func parseSecret(secret apiv1.Secret, authMethod hydrav1alpha1.TokenEndpointAuthMethod) (*hydra.Oauth2ClientCredentials, error) {
	id, found := secret.Data[ClientIDKey]
	if !found {
		return nil, fmt.Errorf("%s property missing", ClientIDKey)
	}

	psw, found := secret.Data[ClientSecretKey]
	if !found && authMethod != "none" {
		return nil, fmt.Errorf("%s property missing", ClientSecretKey)
	}

	return &hydra.Oauth2ClientCredentials{
		ID:       id,
		Password: psw,
	}, nil
}

func (r *OAuth2ClientReconciler) getHydraClientForClient(
	oauth2client hydrav1alpha1.OAuth2Client) (hydra.Client, error) {
	spec := oauth2client.Spec
	if spec.HydraAdmin.URL != "" {
		key := clientKey{
			url:            spec.HydraAdmin.URL,
			port:           spec.HydraAdmin.Port,
			endpoint:       spec.HydraAdmin.Endpoint,
			forwardedProto: spec.HydraAdmin.ForwardedProto,
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if c, ok := r.oauth2Clients[key]; ok {
			return c, nil
		}

		c, err := r.oauth2ClientFactory(spec, "", false)
		if err != nil {
			return nil, fmt.Errorf("cannot create oauth2 c from CRD: %w", err)
		}

		r.oauth2Clients[key] = c
		return c, nil
	}

	if r.HydraClient == nil {
		return nil, fmt.Errorf("no default client configured")
	}

	r.Log.Info("Using default client")

	return r.HydraClient, nil

}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
