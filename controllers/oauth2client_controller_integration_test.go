// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/controllers"
	mocks "github.com/ory/hydra-maester/controllers/mocks/hydra"
	"github.com/ory/hydra-maester/hydra"
)

const (
	timeout      = time.Second * 5
	tstNamespace = "default"
	tstSecret    = "testSecret"
)

var _ = Describe("OAuth2Client Controller", func() {

	Context("in a happy-path scenario", func() {

		Context("should call create OAuth2 client and", func() {

			It("create a Secret if it does not exist", func() {

				tstName, tstClientID, tstSecretName := "test", "testClientID", "my-secret-123"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{
					Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8080",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := &mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					return &hydra.OAuth2ClientJSON{
						ClientID:      &tstClientID,
						Secret:        ptr.To(tstSecret),
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Scope:         o.Scope,
						Audience:      o.Audience,
						Owner:         o.Owner,
					}
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				instance := testInstance(tstName, tstSecretName)
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError.Code).To(BeEmpty())
				Expect(retrieved.Status.ReconciliationError.Description).To(BeEmpty())

				//Verify the created Secret
				var createdSecret apiv1.Secret
				ok = client.ObjectKey{Name: tstSecretName, Namespace: tstNamespace}
				err = k8sClient.Get(context.TODO(), ok, &createdSecret)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdSecret.Data[controllers.ClientIDKey]).To(Equal([]byte(tstClientID)))
				Expect(createdSecret.Data[controllers.ClientSecretKey]).To(Equal([]byte(tstSecret)))
				Expect(createdSecret.OwnerReferences).To(Equal(getOwnerReferenceTo(retrieved)))

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("update object status if the call failed", func() {

				tstName, tstSecretName := "test2", "my-secret-456"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8081",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := &mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("PostOAuth2Client", Anything).Return(nil, errors.New("error"))
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				instance := testInstance(tstName, tstSecretName)
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError).NotTo(BeNil())

				Expect(retrieved.Status.ReconciliationError.Code).To(Equal(hydrav1alpha1.StatusRegistrationFailed))
				Expect(retrieved.Status.ReconciliationError.Description).To(Equal("error"))

				//Verify no secret has been created
				var createdSecret apiv1.Secret
				ok = client.ObjectKey{Name: tstSecretName, Namespace: tstNamespace}
				err = k8sClient.Get(context.TODO(), ok, &createdSecret)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("use provided Secret if it exists", func() {

				tstName, tstClientID, tstSecretName := "test3", "testClientID-3", "my-secret-789"
				var postedClient *hydra.OAuth2ClientJSON
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8082",
					}})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					postedClient = &hydra.OAuth2ClientJSON{
						ClientID:      o.ClientID,
						Secret:        o.Secret,
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Audience:      o.Audience,
						Scope:         o.Scope,
						Owner:         o.Owner,
					}
					return postedClient
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				//ensure secret exists
				secret := apiv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tstSecretName,
						Namespace: tstNamespace,
					},
					Data: map[string][]byte{
						controllers.ClientIDKey:     []byte(tstClientID),
						controllers.ClientSecretKey: []byte(tstSecret),
					},
				}
				err = c.Create(context.TODO(), &secret)
				Expect(err).NotTo(HaveOccurred())

				instance := testInstance(tstName, tstSecretName)
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError.Code).To(BeEmpty())
				Expect(retrieved.Status.ReconciliationError.Description).To(BeEmpty())

				// Ensure that secret doesn't have OwnerReference set
				ok = client.ObjectKey{Name: tstSecretName, Namespace: tstNamespace}
				err = k8sClient.Get(context.TODO(), ok, &secret)
				Expect(err).To(BeNil())
				Expect(len(secret.OwnerReferences)).To(Equal(0))

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("update object status if provided Secret is invalid", func() {

				tstName, tstClientID, tstSecretName := "test4", "testClientID-4", "my-secret-000"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{Scheme: s, Metrics: server.Options{
					BindAddress: ":8083",
				}})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				//ensure invalid secret exists
				secret := apiv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tstSecretName,
						Namespace: tstNamespace,
					},
					Data: map[string][]byte{
						controllers.ClientIDKey: []byte(tstClientID),
						//missing client secret key
					},
				}
				err = c.Create(context.TODO(), &secret)
				Expect(err).NotTo(HaveOccurred())

				instance := testInstance(tstName, tstSecretName)
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError).NotTo(BeNil())
				Expect(retrieved.Status.ReconciliationError.Code).To(Equal(hydrav1alpha1.StatusInvalidSecret))
				Expect(retrieved.Status.ReconciliationError.Description).To(Equal("CLIENT_SECRET property missing"))

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("tolerate nil client_secret if tokenEndpointAuthMethod is none", func() {
				tstName, tstClientID, tstSecretName := "test5", "testClientID-5", "my-secret-without-client-secret"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{Scheme: s, Metrics: server.Options{
					BindAddress: ":8085",
				}})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				mch := &mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					return &hydra.OAuth2ClientJSON{
						ClientID:      &tstClientID,
						Secret:        nil,
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Scope:         o.Scope,
						Audience:      o.Audience,
						Owner:         o.Owner,
					}
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))

				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				instance := testInstance(tstName, tstSecretName)
				instance.Spec.TokenEndpointAuthMethod = "none"
				err = c.Create(context.TODO(), instance)
				// The instance object may not be a valid object because it might be missing some required fields.
				// Please modify the instance object by adding required fields and then remove the following if statement.
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				//Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError.Code).To(BeEmpty())
				Expect(retrieved.Status.ReconciliationError.Description).To(BeEmpty())

				//Verify the created Secret
				var createdSecret apiv1.Secret
				ok = client.ObjectKey{Name: tstSecretName, Namespace: tstNamespace}
				err = k8sClient.Get(context.TODO(), ok, &createdSecret)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdSecret.Data[controllers.ClientIDKey]).To(Equal([]byte(tstClientID)))
				Expect(createdSecret.Data[controllers.ClientSecretKey]).To(BeNil())
				Expect(createdSecret.OwnerReferences).To(Equal(getOwnerReferenceTo(retrieved)))

				//delete instance
				c.Delete(context.TODO(), instance)

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("not delete OAuth2 clients with Orphan deletion policy", func() {
				tstName, tstClientID, tstSecretName := "test-orphan", "testClientID-orphan", "my-secret-orphan"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{
					Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8086",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				deleteHasHappened := false
				mch := &mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(func(id string) error {
					deleteHasHappened = true
					return nil
				})
				mch.On("ListOAuth2Client", Anything).Return(func() []*hydra.OAuth2ClientJSON {
					return []*hydra.OAuth2ClientJSON{
						{
							ClientID: &tstClientID,
							Secret:   ptr.To(tstSecret),
							Owner:    fmt.Sprintf("%s/%s", tstName, tstNamespace),
						},
					}
				}, nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					return &hydra.OAuth2ClientJSON{
						ClientID:      &tstClientID,
						Secret:        ptr.To(tstSecret),
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Scope:         o.Scope,
						Audience:      o.Audience,
						Owner:         o.Owner,
					}
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))
				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				// Create OAuth2 client with 'Orphan' policy
				instance := testInstance(tstName, tstSecretName)
				instance.Spec.DeletionPolicy = hydrav1alpha1.OAuth2ClientDeletionPolicyOrphan

				// Call creation API, to actually create the CRD.
				err = c.Create(context.TODO(), instance)
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				// Call deletion API, which should not really delete the CRD because we are in orphan mode.
				err = c.Delete(context.TODO(), instance)
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to delete object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				Expect(deleteHasHappened).To(BeFalse())

				//Ensure manager is stopped properly
				stopMgr.Done()
			})

			It("delete OAuth2 clients with Delete deletion policy", func() {
				tstName, tstClientID, tstSecretName := "test-delete", "testClientID-delete", "my-secret-delete"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
				// channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{
					Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8087",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				deleteHasHappened := false
				mch := &mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("DeleteOAuth2Client", AnythingOfType("string")).Return(func(id string) error {
					deleteHasHappened = true
					return nil
				})
				mch.On("ListOAuth2Client", Anything).Return(func() []*hydra.OAuth2ClientJSON {
					return []*hydra.OAuth2ClientJSON{
						{
							ClientID: &tstClientID,
							Secret:   ptr.To(tstSecret),
							Owner:    fmt.Sprintf("%s/%s", tstName, tstNamespace),
						},
					}
				}, nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					return &hydra.OAuth2ClientJSON{
						ClientID:      &tstClientID,
						Secret:        ptr.To(tstSecret),
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Scope:         o.Scope,
						Audience:      o.Audience,
						Owner:         o.Owner,
					}
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))
				Expect(add(mgr, recFn)).To(Succeed())

				//Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				// Create OAuth2 client with 'Delete' policy
				instance := testInstance(tstName, tstSecretName)
				instance.Spec.DeletionPolicy = hydrav1alpha1.OAuth2ClientDeletionPolicyDelete

				// Call creation API, to actually create the CRD.
				err = c.Create(context.TODO(), instance)
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				// Call deletion API, which should really delete the CRD because we are in orphan mode.
				err = c.Delete(context.TODO(), instance)
				if apierrors.IsInvalid(err) {
					Fail(fmt.Sprintf("failed to delete object, got an invalid object error: %v", err))
					return
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				Expect(deleteHasHappened).To(BeTrue())

				// Ensure manager is stopped properly.
				stopMgr.Done()
			})

			It("create OAuth2 client with the provided credentials", func() {
				tstName, tstClientID, tstSecretName := "test-create-with-credentials", "test-client-id-with-credentials", "my-secret-with-credentials"
				expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

				s := runtime.NewScheme()
				err := hydrav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				err = apiv1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				// Setup the Manager and Controller.
				// Wrap the Controller Reconcile function so it writes each request to a channel when it is finished.
				mgr, err := manager.New(cfg, manager.Options{
					Scheme: s,
					Metrics: server.Options{
						BindAddress: ":8088",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				c := mgr.GetClient()

				var createdClient *hydra.OAuth2ClientJSON
				mch := mocks.Client{}
				mch.On("GetOAuth2Client", Anything).Return(nil, false, nil)
				mch.On("ListOAuth2Client", Anything).Return(nil, nil)
				mch.On("DeleteOAuth2Client", Anything).Return(nil)
				mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
					createdClient = &hydra.OAuth2ClientJSON{
						ClientID:      o.ClientID,
						Secret:        o.Secret,
						GrantTypes:    o.GrantTypes,
						ResponseTypes: o.ResponseTypes,
						RedirectURIs:  o.RedirectURIs,
						Audience:      o.Audience,
						Scope:         o.Scope,
						Owner:         o.Owner,
					}
					return createdClient
				}, func(o *hydra.OAuth2ClientJSON) error {
					return nil
				})

				recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))
				Expect(add(mgr, recFn)).To(Succeed())

				// Start the manager and the controller
				stopMgr := StartTestManager(mgr)

				// Create the secret
				secret := apiv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tstSecretName,
						Namespace: tstNamespace,
					},
					Data: map[string][]byte{
						controllers.ClientIDKey:     []byte(tstClientID),
						controllers.ClientSecretKey: []byte(tstSecret),
					},
				}
				err = c.Create(context.TODO(), &secret)
				Expect(err).NotTo(HaveOccurred())

				instance := testInstance(tstName, tstSecretName)
				err = c.Create(context.TODO(), instance)
				Expect(err).NotTo(HaveOccurred())
				Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))

				// Verify the created CR instance status
				var retrieved hydrav1alpha1.OAuth2Client
				ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
				err = c.Get(context.TODO(), ok, &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrieved.Status.ReconciliationError.Code).To(BeEmpty())
				Expect(retrieved.Status.ReconciliationError.Description).To(BeEmpty())

				// Ensure the created client has the expected client ID and client secret
				Expect(createdClient.ClientID).ShouldNot(BeNil())
				Expect(createdClient.Secret).ShouldNot(BeNil())
				Expect(*createdClient.ClientID).To(Equal(tstClientID))
				Expect(*createdClient.Secret).To(Equal(tstSecret))

				// Delete instance
				c.Delete(context.TODO(), instance)

				// Ensure manager is stopped properly
				stopMgr.Done()
			})
		})
	})
})

func getOwnerReferenceTo(c hydrav1alpha1.OAuth2Client) []metav1.OwnerReference {
	return []metav1.OwnerReference{{
		APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Name:       c.Name,
		UID:        c.UID,
	}}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Api
	err = c.Watch(source.Kind(mgr.GetCache(), &hydrav1alpha1.OAuth2Client{}, &handler.TypedEnqueueRequestForObject[*hydrav1alpha1.OAuth2Client]{}))
	if err != nil {
		return err
	}

	return nil
}

func getAPIReconciler(mgr ctrl.Manager, mock hydra.Client) reconcile.Reconciler {
	clientMocker := func(spec hydrav1alpha1.OAuth2ClientSpec, tlsTrustStore string, insecureSkipVerify bool) (hydra.Client, error) {
		return mock, nil
	}

	return controllers.New(
		mgr.GetClient(),
		mock,
		ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
		controllers.WithClientFactory(clientMocker),
	)
}

func testInstance(name, secretName string) *hydrav1alpha1.OAuth2Client {

	return &hydrav1alpha1.OAuth2Client{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tstNamespace,
		},
		Spec: hydrav1alpha1.OAuth2ClientSpec{
			GrantTypes:             []hydrav1alpha1.GrantType{"client_credentials"},
			ResponseTypes:          []hydrav1alpha1.ResponseType{"token"},
			Scope:                  "a b c",
			RedirectURIs:           []hydrav1alpha1.RedirectURI{"https://example.com"},
			PostLogoutRedirectURIs: []hydrav1alpha1.RedirectURI{"https://example.com/logout"},
			Audience:               []string{"audience-a"},
			SecretName:             secretName,
			HydraAdmin: hydrav1alpha1.HydraAdmin{
				URL:            "http://hydra-admin",
				Port:           4445,
				Endpoint:       "/client",
				ForwardedProto: "https",
			},
		}}
}
