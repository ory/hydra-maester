package controllers_test

import (
	"time"
)

const (
	timeout       = time.Second * 5
	tstNamespace  = "default"
	tstScopes     = "a b c"
	tstSecretName = "my-secret-123"
)

//var tstClientID = "testClientID"
//var tstSecret = "testSecret"
//
//var _ = Describe("OAuth2Client Controller", func() {
//	Context("in a happy-path scenario", func() {
//
//		It("should call create OAuth2 client in Hydra and a Secret", func() {
//
//			tstName := "test"
//			expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}
//
//			s := scheme.Scheme
//			err := hydrav1alpha1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			err = apiv1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
//			// channel when it is finished.
//			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
//			Expect(err).NotTo(HaveOccurred())
//			c := mgr.GetClient()
//
//			mch := mocks.HydraClientInterface{}
//			mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
//				return &hydra.OAuth2ClientJSON{
//					ClientID:      &tstClientID,
//					Secret:        &tstSecret,
//					Name:          o.Name,
//					GrantTypes:    o.GrantTypes,
//					ResponseTypes: o.ResponseTypes,
//					Scope:         o.Scope,
//				}
//			}, func(o *hydra.OAuth2ClientJSON) error {
//				return nil
//			})
//
//			recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))
//
//			Expect(add(mgr, recFn)).To(Succeed())
//
//			//Start the manager and the controller
//			stopMgr, mgrStopped := StartTestManager(mgr)
//
//			//Ensure manager is stopped properly
//			defer func() {
//				close(stopMgr)
//				mgrStopped.Wait()
//			}()
//
//			instance := testInstance(tstName)
//			err = c.Create(context.TODO(), instance)
//			// The instance object may not be a valid object because it might be missing some required fields.
//			// Please modify the instance object by adding required fields and then remove the following if statement.
//			if apierrors.IsInvalid(err) {
//				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
//				return
//			}
//			Expect(err).NotTo(HaveOccurred())
//			//defer c.Delete(context.TODO(), instance)
//			Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))
//
//			//Verify the created CR instance status
//			var retrieved hydrav1alpha1.OAuth2Client
//			ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
//			err = c.Get(context.TODO(), ok, &retrieved)
//			Expect(err).NotTo(HaveOccurred())
//
//			//Expect(*retrieved.Status.ClientID).To(Equal(tstClientID))
//			//Expect(*retrieved.Status.Secret).To(Equal(tstName)) //Secret contents is not visible in the CR instance!
//
//			//Verify the created Secret
//			var createdSecret = apiv1.Secret{}
//			ok = client.ObjectKey{Name: tstSecretName, Namespace: tstNamespace}
//			k8sClient.Get(context.TODO(), ok, &createdSecret)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(createdSecret.Data[controllers.ClientIDKey]).To(Equal([]byte(tstClientID)))
//			Expect(createdSecret.Data[controllers.ClientSecretKey]).To(Equal([]byte(tstSecret)))
//		})
//
//		It("should call create OAuth2 client in Hydra and update object status if the call failed", func() {
//
//			tstName := "test2"
//			expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}
//
//			s := scheme.Scheme
//			err := hydrav1alpha1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			err = apiv1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
//			// channel when it is finished.
//			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
//			Expect(err).NotTo(HaveOccurred())
//			c := mgr.GetClient()
//
//			mch := mocks.HydraClientInterface{}
//			mch.On("PostOAuth2Client", Anything).Return(nil, errors.New("error"))
//
//			recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))
//
//			Expect(add(mgr, recFn)).To(Succeed())
//
//			//Start the manager and the controller
//			stopMgr, mgrStopped := StartTestManager(mgr)
//
//			//Ensure manager is stopped properly
//			defer func() {
//				close(stopMgr)
//				mgrStopped.Wait()
//			}()
//
//			instance := testInstance(tstName)
//			err = c.Create(context.TODO(), instance)
//			// The instance object may not be a valid object because it might be missing some required fields.
//			// Please modify the instance object by adding required fields and then remove the following if statement.
//			if apierrors.IsInvalid(err) {
//				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
//				return
//			}
//			Expect(err).NotTo(HaveOccurred())
//			//defer c.Delete(context.TODO(), instance)
//			Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))
//
//			//Verify the created CR instance status
//			var retrieved hydrav1alpha1.OAuth2Client
//			ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
//			err = c.Get(context.TODO(), ok, &retrieved)
//			Expect(err).NotTo(HaveOccurred())
//
//			//Expect(retrieved.Status.ClientID).To(BeNil())
//			//Expect(retrieved.Status.Secret).To(BeNil())
//			Expect(retrieved.Status.ReconciliationError).NotTo(BeNil())
//			Expect(retrieved.Status.ReconciliationError.Code).To(Equal(hydrav1alpha1.StatusRegistrationFailed))
//			Expect(retrieved.Status.ReconciliationError.Description).To(Equal("error"))
//
//		})
//
//		tstClientID = "testClientID2"
//
//		It("should call create OAuth2 client in Hydra, create Secret and update object status if Secret creation failed", func() {
//
//			tstName := "test3"
//			expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}
//
//			s := scheme.Scheme
//			err := hydrav1alpha1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			err = apiv1.AddToScheme(s)
//			Expect(err).NotTo(HaveOccurred())
//
//			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
//			// channel when it is finished.
//			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
//			Expect(err).NotTo(HaveOccurred())
//			c := mgr.GetClient()
//
//			mch := mocks.HydraClientInterface{}
//			mch.On("PostOAuth2Client", AnythingOfType("*hydra.OAuth2ClientJSON")).Return(func(o *hydra.OAuth2ClientJSON) *hydra.OAuth2ClientJSON {
//				return &hydra.OAuth2ClientJSON{
//					ClientID:      &tstClientID,
//					Secret:        &tstSecret,
//					Name:          o.Name,
//					GrantTypes:    o.GrantTypes,
//					ResponseTypes: o.ResponseTypes,
//					Scope:         o.Scope,
//				}
//			}, func(o *hydra.OAuth2ClientJSON) error {
//				return nil
//			})
//
//			recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, &mch))
//
//			Expect(add(mgr, recFn)).To(Succeed())
//
//			//Start the manager and the controller
//			stopMgr, mgrStopped := StartTestManager(mgr)
//
//			//Ensure manager is stopped properly
//			defer func() {
//				close(stopMgr)
//				mgrStopped.Wait()
//			}()
//
//			//ensure conflicting entry exists
//			secret := apiv1.Secret{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      tstSecretName,
//					Namespace: tstNamespace,
//				},
//			}
//			err = c.Create(context.TODO(), &secret)
//			Expect(err).NotTo(HaveOccurred())
//
//			instance := testInstance(tstName)
//			err = c.Create(context.TODO(), instance)
//			// The instance object may not be a valid object because it might be missing some required fields.
//			// Please modify the instance object by adding required fields and then remove the following if statement.
//			if apierrors.IsInvalid(err) {
//				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
//				return
//			}
//			Expect(err).NotTo(HaveOccurred())
//			//defer c.Delete(context.TODO(), instance)
//			Eventually(requests, timeout).Should(Receive(Equal(*expectedRequest)))
//
//			//Verify the created CR instance status
//			var retrieved hydrav1alpha1.OAuth2Client
//			ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
//			err = c.Get(context.TODO(), ok, &retrieved)
//			Expect(err).NotTo(HaveOccurred())
//
//			//Expect(retrieved.Status.ClientID).NotTo(BeNil())
//			//Expect(retrieved.Status.Secret).To(BeNil())
//			Expect(retrieved.Status.ReconciliationError).NotTo(BeNil())
//			Expect(retrieved.Status.ReconciliationError.Code).To(Equal(hydrav1alpha1.StatusCreateSecretFailed))
//			Expect(retrieved.Status.ReconciliationError.Description).To(Equal(`secrets "test3" already exists`))
//		})
//	})
//})
//
//// add adds a new Controller to mgr with r as the reconcile.Reconciler
//func add(mgr manager.Manager, r reconcile.Reconciler) error {
//	// Create a new controller
//	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
//	if err != nil {
//		return err
//	}
//
//	// Watch for changes to Api
//	err = c.Watch(&source.Kind{Type: &hydrav1alpha1.OAuth2Client{}}, &handler.EnqueueRequestForObject{})
//	if err != nil {
//		return err
//	}
//
//	// TODO(user): Modify this to be the types you create
//	// Uncomment watch a Deployment created by Guestbook - change this for objects you create
//	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
//	//	IsController: true,
//	//	OwnerType:    &webappv1.Guestbook{},
//	//})
//	//if err != nil {
//	//	return err
//	//}
//
//	return nil
//}
//
//func getAPIReconciler(mgr ctrl.Manager, mock controllers.HydraClientInterface) reconcile.Reconciler {
//	return &controllers.OAuth2ClientReconciler{
//		Client:      mgr.GetClient(),
//		Log:         ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
//		HydraClient: mock,
//	}
//}
//
//func testInstance(name string) *hydrav1alpha1.OAuth2Client {
//
//	return &hydrav1alpha1.OAuth2Client{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      name,
//			Namespace: tstNamespace,
//		},
//		Spec: hydrav1alpha1.OAuth2ClientSpec{
//			GrantTypes:    []hydrav1alpha1.GrantType{"client_credentials"},
//			ResponseTypes: []hydrav1alpha1.ResponseType{"token"},
//			Scope:         tstScopes,
//			SecretName:    tstSecretName,
//		}}
//}
