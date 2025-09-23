// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra_test

import (
	"testing"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/hydra"
	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	t.Run("Test ScopeArray", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				ScopeArray: []string{"scope1", "scope2"},
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		assert.Equal(t, "scope1 scope2", parsedClient.Scope)
	})

	t.Run("Test having both Scope and ScopeArray", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				Scope:      "scope3",
				ScopeArray: []string{"scope1", "scope2"},
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		assert.Equal(t, "scope1 scope2 scope3", parsedClient.Scope)
	})

	t.Run("Test having jwks uri", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				JwksUri: "https://ory.sh/jwks.json",
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		assert.Equal(t, "https://ory.sh/jwks.json", parsedClient.JwksUri)
	})

	t.Run("Test jwks uri is required when token endpoint auth method is private_key_jwt", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				TokenEndpointAuthMethod: "private_key_jwt",
			},
		}

		var _, err = hydra.FromOAuth2Client(&c)

		assert.ErrorContains(t, err, "JwksUri")
	})

	t.Run("Test RequestURIs conversion", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				RequestURIs: []hydrav1alpha1.RedirectURI{
					"https://example.com/request1",
					"https://example.com/request2",
				},
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		expected := []string{"https://example.com/request1", "https://example.com/request2"}
		assert.Equal(t, expected, parsedClient.RequestURIs)
	})

	t.Run("Test AccessTokenStrategy field", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				AccessTokenStrategy: "jwt",
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		assert.Equal(t, "jwt", parsedClient.AccessTokenStrategy)
	})

	t.Run("Test multiple new fields", func(t *testing.T) {
		c := hydrav1alpha1.OAuth2Client{
			Spec: hydrav1alpha1.OAuth2ClientSpec{
				ClientUri:                   "https://example.com",
				Contacts:                    []string{"admin@example.com", "support@example.com"},
				PolicyUri:                   "https://example.com/privacy",
				TosUri:                      "https://example.com/terms",
				SubjectType:                 "public",
				SkipLogoutConsent:           true,
				ClientSecretExpiresAt:       1234567890,
			},
		}

		var parsedClient, err = hydra.FromOAuth2Client(&c)
		if err != nil {
			assert.Fail(t, "unexpected error: %s", err)
		}

		assert.Equal(t, "https://example.com", parsedClient.ClientUri)
		assert.Equal(t, []string{"admin@example.com", "support@example.com"}, parsedClient.Contacts)
		assert.Equal(t, "https://example.com/privacy", parsedClient.PolicyUri)
		assert.Equal(t, "https://example.com/terms", parsedClient.TosUri)
		assert.Equal(t, "public", parsedClient.SubjectType)
		assert.Equal(t, true, parsedClient.SkipLogoutConsent)
		assert.Equal(t, int64(1234567890), parsedClient.ClientSecretExpiresAt)
	})
}
