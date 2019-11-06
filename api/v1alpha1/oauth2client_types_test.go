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

package v1alpha1

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient                    client.Client
	cfg                          *rest.Config
	testEnv                      *envtest.Environment
	key                          types.NamespacedName
	created, fetched             *OAuth2Client
	createErr, getErr, deleteErr error
)

func TestCreateAPI(t *testing.T) {

	runEnv(t)
	defer stopEnv(t)

	t.Run("should handle an object properly", func(t *testing.T) {

		key = types.NamespacedName{
			Name:      "foo",
			Namespace: "default",
		}

		t.Run("by creating an API object if it meets CRD requirements", func(t *testing.T) {

			resetTestClient()

			createErr = k8sClient.Create(context.TODO(), created)
			require.NoError(t, createErr)

			fetched = &OAuth2Client{}
			getErr = k8sClient.Get(context.TODO(), key, fetched)
			require.NoError(t, getErr)
			assert.Equal(t, created, fetched)

			deleteErr = k8sClient.Delete(context.TODO(), created)
			require.NoError(t, deleteErr)

			getErr = k8sClient.Get(context.TODO(), key, created)
			require.Error(t, getErr)
		})

		t.Run("by failing if the requested object doesn't meet CRD requirements", func(t *testing.T) {

			for desc, modifyClient := range map[string]func(){
				"invalid grant type":    func() { created.Spec.GrantTypes = []GrantType{"invalid"} },
				"invalid response type": func() { created.Spec.ResponseTypes = []ResponseType{"invalid"} },
				"invalid scope":         func() { created.Spec.Scope = "" },
				"missing secret name":   func() { created.Spec.SecretName = "" },
				"invalid redirect URI":  func() { created.Spec.RedirectURIs[1] = "invalid" },
			} {
				t.Run(fmt.Sprintf("case=%s", desc), func(t *testing.T) {

					resetTestClient()
					modifyClient()
					createErr = k8sClient.Create(context.TODO(), created)
					require.Error(t, createErr)
				})
			}
		})
	})
}

func runEnv(t *testing.T) {

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}

	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	cfg, err = testEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)

}

func stopEnv(t *testing.T) {
	err := testEnv.Stop()
	require.NoError(t, err)
}

func resetTestClient() {
	created = &OAuth2Client{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: OAuth2ClientSpec{
			GrantTypes:    []GrantType{"implicit", "client_credentials", "authorization_code", "refresh_token"},
			ResponseTypes: []ResponseType{"id_token", "code", "token"},
			Scope:         "read,write",
			SecretName:    "secret-name",
			RedirectURIs:  []RedirectURI{"https://client/account", "http://localhost:8080/account"},
		},
	}
}
