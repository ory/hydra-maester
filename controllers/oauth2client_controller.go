/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/hydra"
	apiv1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clientIDKey     = "client_id"
	clientSecretKey = "client_secret"
)

type HydraClientInterface interface {
	GetOAuth2Client(id string) (*hydra.OAuth2ClientJSON, bool, error)
	PostOAuth2Client(o *hydra.OAuth2ClientJSON) (*hydra.OAuth2ClientJSON, error)
	PutOAuth2Client(o *hydra.OAuth2ClientJSON) (*hydra.OAuth2ClientJSON, error)
	DeleteOAuth2Client(id string) error
}

// OAuth2ClientReconciler reconciles a OAuth2Client object
type OAuth2ClientReconciler struct {
	HydraClient HydraClientInterface
	Log         logr.Logger
	client.Client
}

// +kubebuilder:rbac:groups=hydra.ory.sh,resources=oauth2clients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hydra.ory.sh,resources=oauth2clients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *OAuth2ClientReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("oauth2client", req.NamespacedName)

	var client hydrav1alpha1.OAuth2Client
	if err := r.Get(ctx, req.NamespacedName, &client); err != nil {
		if apierrs.IsNotFound(err) {
			if err := r.unregisterOAuth2Client(ctx, req.NamespacedName); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if client.Generation != client.Status.ObservedGeneration {

		var registered = false
		var err error

		if client.Status.ClientID != nil {

			_, registered, err = r.HydraClient.GetOAuth2Client(*client.Status.ClientID)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		if !registered {
			return ctrl.Result{}, r.registerOAuth2Client(ctx, &client)
		}

		return ctrl.Result{}, r.updateRegisteredOAuth2Client(&client)
	}

	return ctrl.Result{}, nil
}

func (r *OAuth2ClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hydrav1alpha1.OAuth2Client{}).
		Complete(r)
}

func (r *OAuth2ClientReconciler) registerOAuth2Client(ctx context.Context, client *hydrav1alpha1.OAuth2Client) error {
	created, err := r.HydraClient.PostOAuth2Client(client.ToOAuth2ClientJSON())
	if err != nil {
		return err
	}

	clientSecret := apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      client.Name,
			Namespace: client.Namespace,
		},
		Data: map[string][]byte{
			clientSecretKey: []byte(*created.Secret),
			clientIDKey:     []byte(*created.ClientID),
		},
	}

	err = r.Create(ctx, &clientSecret)
	if err != nil {
		return err
	}

	client.Status.Secret = &clientSecret.Name
	client.Status.ClientID = created.ClientID
	client.Status.ObservedGeneration = client.Generation
	return r.Status().Update(ctx, client)
}

func (r *OAuth2ClientReconciler) unregisterOAuth2Client(ctx context.Context, namespacedName types.NamespacedName) error {
	var sec apiv1.Secret
	if err := r.Get(ctx, namespacedName, &sec); err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("unable to find secret corresponding with client %s/%s. Manual deletion recommended", namespacedName.Name, namespacedName.Namespace))
			return nil
		}
		return err
	}

	return r.HydraClient.DeleteOAuth2Client(string(sec.Data[clientIDKey]))
}

func (r *OAuth2ClientReconciler) updateRegisteredOAuth2Client(client *hydrav1alpha1.OAuth2Client) error {
	_, err := r.HydraClient.PutOAuth2Client(client.ToOAuth2ClientJSON())
	return err
}
