// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

		t.Run("by creating an API object if it meets CRD requirements without optional parameters", func(t *testing.T) {

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

		t.Run("by creating an API object if it meets CRD requirements with optional parameters", func(t *testing.T) {

			resetTestClient()

			created.Spec.RedirectURIs = []RedirectURI{"https://client/account", "http://localhost:8080/account"}
			created.Spec.HydraAdmin = HydraAdmin{
				URL:  "http://localhost",
				Port: 4445,
				// Endpoint:       "/clients",
				ForwardedProto: "https",
			}

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
				"invalid grant type":                                func() { created.Spec.GrantTypes = []GrantType{"invalid"} },
				"invalid response type":                             func() { created.Spec.ResponseTypes = []ResponseType{"invalid", "code"} },
				"invalid composite response type":                   func() { created.Spec.ResponseTypes = []ResponseType{"invalid code", "code id_token"} },
				"missing secret name":                               func() { created.Spec.SecretName = "" },
				"invalid redirect URI":                              func() { created.Spec.RedirectURIs = []RedirectURI{"invalid"} },
				"invalid logout redirect URI":                       func() { created.Spec.PostLogoutRedirectURIs = []RedirectURI{"invalid"} },
				"invalid hydra url":                                 func() { created.Spec.HydraAdmin.URL = "invalid" },
				"invalid hydra port high":                           func() { created.Spec.HydraAdmin.Port = 65536 },
				"invalid hydra endpoint":                            func() { created.Spec.HydraAdmin.Endpoint = "invalid" },
				"invalid hydra forwarded proto":                     func() { created.Spec.HydraAdmin.ForwardedProto = "invalid" },
				"invalid lifespan authorization code access token":  func() { created.Spec.TokenLifespans.AuthorizationCodeGrantAccessTokenLifespan = "invalid" },
				"invalid lifespan authorization code id token":      func() { created.Spec.TokenLifespans.AuthorizationCodeGrantIdTokenLifespan = "invalid" },
				"invalid lifespan authorization code refresh token": func() { created.Spec.TokenLifespans.AuthorizationCodeGrantRefreshTokenLifespan = "invalid" },
				"invalid lifespan client credentials access token":  func() { created.Spec.TokenLifespans.ClientCredentialsGrantAccessTokenLifespan = "invalid" },
				"invalid lifespan implicit access token":            func() { created.Spec.TokenLifespans.ImplicitGrantAccessTokenLifespan = "invalid" },
				"invalid lifespan implicit id token":                func() { created.Spec.TokenLifespans.ImplicitGrantIdTokenLifespan = "invalid" },
				"invalid lifespan jwt bearer access token":          func() { created.Spec.TokenLifespans.JwtBearerGrantAccessTokenLifespan = "invalid" },
				"invalid lifespan refresh token access token":       func() { created.Spec.TokenLifespans.RefreshTokenGrantAccessTokenLifespan = "invalid" },
				"invalid lifespan refresh token id token":           func() { created.Spec.TokenLifespans.RefreshTokenGrantIdTokenLifespan = "invalid" },
				"invalid lifespan refresh token refresh token":      func() { created.Spec.TokenLifespans.RefreshTokenGrantRefreshTokenLifespan = "invalid" },
			} {
				t.Run(fmt.Sprintf("case=%s", desc), func(t *testing.T) {
					resetTestClient()
					modifyClient()
					createErr = k8sClient.Create(context.TODO(), created)
					require.Error(t, createErr)
				})
			}
		})

		t.Run("by creating an object if it passes validation", func(t *testing.T) {
			for desc, modifyClient := range map[string]func(){
				"single response type": func() { created.Spec.ResponseTypes = []ResponseType{"token", "id_token", "code"} },
				"double response type": func() { created.Spec.ResponseTypes = []ResponseType{"id_token token", "code id_token", "code token"} },
				"triple response type": func() { created.Spec.ResponseTypes = []ResponseType{"code id_token token"} },
			} {
				t.Run(fmt.Sprintf("case=%s", desc), func(t *testing.T) {
					resetTestClient()
					modifyClient()
					require.NoError(t, k8sClient.Create(context.TODO(), created))
					require.NoError(t, k8sClient.Delete(context.TODO(), created))
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
			GrantTypes:     []GrantType{"implicit", "client_credentials", "authorization_code", "refresh_token"},
			ResponseTypes:  []ResponseType{"id_token", "code", "token"},
			Scope:          "read,write",
			SecretName:     "secret-name",
			TokenLifespans: TokenLifespans{},
		},
	}
}
