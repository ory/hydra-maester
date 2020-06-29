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
	"encoding/json"
	"fmt"

	"github.com/ory/hydra-maester/hydra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusCode string

const (
	StatusRegistrationFailed  StatusCode = "CLIENT_REGISTRATION_FAILED"
	StatusCreateSecretFailed  StatusCode = "SECRET_CREATION_FAILED"
	StatusUpdateFailed        StatusCode = "CLIENT_UPDATE_FAILED"
	StatusInvalidSecret       StatusCode = "INVALID_SECRET"
	StatusInvalidHydraAddress StatusCode = "INVALID_HYDRA_ADDRESS"
)

// HydraAdmin defines the desired hydra admin instance to use for OAuth2Client
type HydraAdmin struct {
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=(^$|^https?://.*)
	//
	// URL is the URL for the hydra instance on
	// which to set up the client. This value will override the value
	// provided to `--hydra-url`
	URL string `json:"url,omitempty"`

	// +kubebuilder:validation:Maximum=65535
	//
	// Port is the port for the hydra instance on
	// which to set up the client. This value will override the value
	// provided to `--hydra-port`
	Port int `json:"port,omitempty"`

	// +kubebuilder:validation:Pattern=(^$|^/.*)
	//
	// Endpoint is the endpoint for the hydra instance on which
	// to set up the client. This value will override the value
	// provided to `--endpoint` (defaults to `"/clients"` in the
	// application)
	Endpoint string `json:"endpoint,omitempty"`

	// +kubebuilder:validation:Pattern=(^$|https?|off)
	//
	// ForwardedProto overrides the `--forwarded-proto` flag. The
	// value "off" will force this to be off even if
	// `--forwarded-proto` is specified
	ForwardedProto string `json:"forwardedProto,omitempty"`
}

// OAuth2ClientSpec defines the desired state of OAuth2Client
type OAuth2ClientSpec struct {

	// ClientName is the human-readable string name of the client to be presented to the end-user during authorization.
	ClientName string `json:"clientName,omitempty"`

	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:MinItems=1
	//
	// GrantTypes is an array of grant types the client is allowed to use.
	GrantTypes []GrantType `json:"grantTypes"`

	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:MinItems=1
	//
	// ResponseTypes is an array of the OAuth 2.0 response type strings that the client can
	// use at the authorization endpoint.
	ResponseTypes []ResponseType `json:"responseTypes,omitempty"`

	// RedirectURIs is an array of the redirect URIs allowed for the application
	RedirectURIs []RedirectURI `json:"redirectUris,omitempty"`

	// PostLogoutRedirectURIs is an array of the post logout redirect URIs allowed for the application
	PostLogoutRedirectURIs []RedirectURI `json:"postLogoutRedirectUris,omitempty"`

	// AllowedCorsOrigins is an array of allowed CORS origins
	AllowedCorsOrigins []RedirectURI `json:"allowedCorsOrigins,omitempty"`

	// Audience is a whitelist defining the audiences this client is allowed to request tokens for
	Audience []string `json:"audience,omitempty"`

	// +kubebuilder:validation:Pattern=([a-zA-Z0-9\.\*]+\s?)+
	//
	// Scope is a string containing a space-separated list of scope values (as
	// described in Section 3.3 of OAuth 2.0 [RFC6749]) that the client
	// can use when requesting access tokens.
	Scope string `json:"scope"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	//
	// SecretName points to the K8s secret that contains this client's ID and password
	SecretName string `json:"secretName"`

	// HydraAdmin is the optional configuration to use for managing
	// this client
	HydraAdmin HydraAdmin `json:"hydraAdmin,omitempty"`

	// +kubebuilder:validation:Enum=client_secret_basic;client_secret_post;private_key_jwt;none
	//
	// Indication which authentication method shoud be used for the token endpoint
	TokenEndpointAuthMethod TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`

	// Metadata is abritrary data
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// +kubebuilder:validation:Enum=client_credentials;authorization_code;implicit;refresh_token
// GrantType represents an OAuth 2.0 grant type
type GrantType string

// +kubebuilder:validation:Enum=id_token;code;token
// ResponseType represents an OAuth 2.0 response type strings
type ResponseType string

// +kubebuilder:validation:Pattern=\w+:/?/?[^\s]+
// RedirectURI represents a redirect URI for the client
type RedirectURI string

// +kubebuilder:validation:Enum=client_secret_basic;client_secret_post;private_key_jwt;none
// TokenEndpointAuthMethod represents an authentication method for token endpoint
type TokenEndpointAuthMethod string

// OAuth2ClientStatus defines the observed state of OAuth2Client
type OAuth2ClientStatus struct {
	// ObservedGeneration represents the most recent generation observed by the daemon set controller.
	ObservedGeneration  int64               `json:"observedGeneration,omitempty"`
	ReconciliationError ReconciliationError `json:"reconciliationError,omitempty"`
}

// ReconciliationError represents an error that occurred during the reconciliation process
type ReconciliationError struct {
	// Code is the status code of the reconciliation error
	Code StatusCode `json:"statusCode,omitempty"`
	// Description is the description of the reconciliation error
	Description string `json:"description,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OAuth2Client is the Schema for the oauth2clients API
type OAuth2Client struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OAuth2ClientSpec   `json:"spec,omitempty"`
	Status OAuth2ClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OAuth2ClientList contains a list of OAuth2Client
type OAuth2ClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OAuth2Client `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OAuth2Client{}, &OAuth2ClientList{})
}

// ToOAuth2ClientJSON converts an OAuth2Client into a OAuth2ClientJSON object that represents an OAuth2 client digestible by ORY Hydra
func (c *OAuth2Client) ToOAuth2ClientJSON() *hydra.OAuth2ClientJSON {
	return &hydra.OAuth2ClientJSON{
		ClientName:              c.Spec.ClientName,
		GrantTypes:              grantToStringSlice(c.Spec.GrantTypes),
		ResponseTypes:           responseToStringSlice(c.Spec.ResponseTypes),
		RedirectURIs:            redirectToStringSlice(c.Spec.RedirectURIs),
		PostLogoutRedirectURIs:  redirectToStringSlice(c.Spec.PostLogoutRedirectURIs),
		AllowedCorsOrigins:      redirectToStringSlice(c.Spec.AllowedCorsOrigins),
		Audience:                c.Spec.Audience,
		Scope:                   c.Spec.Scope,
		Owner:                   fmt.Sprintf("%s/%s", c.Name, c.Namespace),
		TokenEndpointAuthMethod: string(c.Spec.TokenEndpointAuthMethod),
		Metadata:                c.Spec.Metadata,
	}
}

func responseToStringSlice(rt []ResponseType) []string {
	var output = make([]string, len(rt))
	for i, elem := range rt {
		output[i] = string(elem)
	}
	return output
}

func grantToStringSlice(gt []GrantType) []string {
	var output = make([]string, len(gt))
	for i, elem := range gt {
		output[i] = string(elem)
	}
	return output
}

func redirectToStringSlice(ru []RedirectURI) []string {
	var output = make([]string, len(ru))
	for i, elem := range ru {
		output[i] = string(elem)
	}
	return output
}
