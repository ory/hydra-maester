package controllers_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/controllers"
	"github.com/ory/hydra-maester/hydra"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const timeout = time.Second * 5

var _ = Describe("OAuth2Client Controller", func() {
	Context("in a happy-path scenario", func() {

		var tstName = "test"
		var tstNamespace = "default"
		var tstScopes = "a b c"
		var tstClientID = "testClientID"
		var tstSecret = "testSecret"

		var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}
		It("should call create OAuth2 client in Hydra and a Secret", func() {

			s := scheme.Scheme
			err := hydrav1alpha1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
			// channel when it is finished.
			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
			Expect(err).NotTo(HaveOccurred())
			c := mgr.GetClient()

			mch := (&mockHydraClient{}).
				withSecret(tstSecret).
				withClientID(tstClientID)

			recFn, requests := SetupTestReconcile(getAPIReconciler(mgr, mch))
			//_, requests := SetupTestReconcile(getApiReconciler(mgr))

			Expect(add(mgr, recFn)).To(Succeed())

			//Start the manager and the controller
			stopMgr, mgrStopped := StartTestManager(mgr)

			//Ensure manager is stopped properly
			defer func() {
				close(stopMgr)
				mgrStopped.Wait()
			}()

			instance := testInstance(tstName, tstNamespace, tstScopes)
			err = c.Create(context.TODO(), instance)
			// The instance object may not be a valid object because it might be missing some required fields.
			// Please modify the instance object by adding required fields and then remove the following if statement.
			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer c.Delete(context.TODO(), instance)
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			//Verify the created CR instance status
			var retrieved hydrav1alpha1.OAuth2Client
			ok := client.ObjectKey{Name: tstName, Namespace: tstNamespace}
			err = c.Get(context.TODO(), ok, &retrieved)
			Expect(err).NotTo(HaveOccurred())

			Expect(*retrieved.Status.ClientID).To(Equal(tstClientID))
			Expect(*retrieved.Status.Secret).To(Equal(tstName)) //Secret contents is not visible in the CR instance!

			//Verify the created Secret
			var createdSecret = apiv1.Secret{}
			k8sClient.Get(context.TODO(), ok, &createdSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdSecret.Data["client_secret"]).To(Equal([]byte(tstSecret)))
		})
	})
})

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Api
	err = c.Watch(&source.Kind{Type: &hydrav1alpha1.OAuth2Client{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Guestbook - change this for objects you create
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &webappv1.Guestbook{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

func getAPIReconciler(mgr ctrl.Manager, mock *mockHydraClient) reconcile.Reconciler {
	return &controllers.OAuth2ClientReconciler{
		Client:      mgr.GetClient(),
		Log:         ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
		HydraClient: mock,
	}
}

func testInstance(name, namespace, scopes string) *hydrav1alpha1.OAuth2Client {

	return &hydrav1alpha1.OAuth2Client{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: hydrav1alpha1.OAuth2ClientSpec{
			GrantTypes:    []hydrav1alpha1.GrantType{"client_credentials"},
			ResponseTypes: []hydrav1alpha1.ResponseType{"token"},
			Scope:         scopes,
		}}
}

//TODO: Replace with full-fledged mocking framework (mockery/go-mock)
type mockHydraClient struct {
	resSecret   string
	resClientID string
	postedData  *hydra.OAuth2ClientJSON
}

func (m *mockHydraClient) withSecret(secret string) *mockHydraClient {
	m.resSecret = secret
	return m
}

func (m *mockHydraClient) withClientID(clientID string) *mockHydraClient {
	m.resClientID = clientID
	return m
}

//Returns the data previously "stored" by PostOAuth2Client
func (m *mockHydraClient) GetOAuth2Client(id string) (*hydra.OAuth2ClientJSON, bool, error) {
	res := &hydra.OAuth2ClientJSON{
		ClientID:      &m.resClientID,
		Secret:        &m.resSecret,
		Name:          m.postedData.Name,
		GrantTypes:    m.postedData.GrantTypes,
		ResponseTypes: m.postedData.ResponseTypes,
		Scope:         m.postedData.Scope,
	}
	return res, true, nil
}

func (m *mockHydraClient) PostOAuth2Client(o *hydra.OAuth2ClientJSON) (*hydra.OAuth2ClientJSON, error) {
	m.postedData = o
	res := &hydra.OAuth2ClientJSON{
		ClientID:      &m.resClientID,
		Secret:        &m.resSecret,
		Name:          o.Name,
		GrantTypes:    o.GrantTypes,
		ResponseTypes: o.ResponseTypes,
		Scope:         o.Scope,
	}
	return res, nil
}

func (m *mockHydraClient) DeleteOAuth2Client(id string) error {
	panic("Should not be invoked!")
}

func (m *mockHydraClient) PutOAuth2Client(o *hydra.OAuth2ClientJSON) (*hydra.OAuth2ClientJSON, error) {
	panic("Should not be invoked!")
}
