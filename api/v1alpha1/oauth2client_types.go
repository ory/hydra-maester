// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	// +kubebuilder:validation:Pattern=`(^$|^https?://.*)`
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

	// ApiKeySecretRef is an object to define the secret which contains
	// Ory Network API Key
	ApiKeySecretRef ApiKeySecretRef `json:"apiKeySecretRef,omitempty"`
}

// ApiKeySecretRef contains Secret details for the API Key
type ApiKeySecretRef struct {
	// Name of the secret containing the API Key
	Name string `json:"name,omitempty"`
	// Key of the secret for the API key
	Key string `json:"key,omitempty"`
	// Namespace of the secret if different from hydra-maester controller
	Namespace string `json:"namespace,omitempty"`
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

	// SkipConsent skips the consent screen for this client.
	// +kubebuilder:validation:type=bool
	// +kubebuilder:default=false
	SkipConsent bool `json:"skipConsent,omitempty"`

	// HydraAdmin is the optional configuration to use for managing
	// this client
	HydraAdmin HydraAdmin `json:"hydraAdmin,omitempty"`

	// +kubebuilder:validation:Enum=client_secret_basic;client_secret_post;private_key_jwt;none
	//
	// Indication which authentication method shoud be used for the token endpoint
	TokenEndpointAuthMethod TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`

	// +kubebuilder:validation:Type=object
	// +nullable
	// +optional
	//
	// Metadata is abritrary data
	Metadata apiextensionsv1.JSON `json:"metadata,omitempty"`

	// +kubebuilder:validation:type=string
	// +kubebuilder:validation:Pattern=`(^$|^https?://.*)`
	//
	// JwksUri Define the URL where the JSON Web Key Set should be fetched from when performing the private_key_jwt client authentication method.
	JwksUri string `json:"jwksUri,omitempty"`
}

// +kubebuilder:validation:Enum=client_credentials;authorization_code;implicit;refresh_token
// GrantType represents an OAuth 2.0 grant type
type GrantType string

// +kubebuilder:validation:Enum=id_token;code;token;code token;code id_token;id_token token;code id_token token
// ResponseType represents an OAuth 2.0 response type strings
type ResponseType string

// +kubebuilder:validation:Pattern=`\w+:/?/?[^\s]+`
// RedirectURI represents a redirect URI for the client
type RedirectURI string

// +kubebuilder:validation:Enum=client_secret_basic;client_secret_post;private_key_jwt;none
// TokenEndpointAuthMethod represents an authentication method for token endpoint
type TokenEndpointAuthMethod string

// OAuth2ClientStatus defines the observed state of OAuth2Client
type OAuth2ClientStatus struct {
	// ObservedGeneration represents the most recent generation observed by the daemon set controller.
	ObservedGeneration  int64                   `json:"observedGeneration,omitempty"`
	ReconciliationError ReconciliationError     `json:"reconciliationError,omitempty"`
	Conditions          []OAuth2ClientCondition `json:"conditions,omitempty"`
}

// ReconciliationError represents an error that occurred during the reconciliation process
type ReconciliationError struct {
	// Code is the status code of the reconciliation error
	Code StatusCode `json:"statusCode,omitempty"`
	// Description is the description of the reconciliation error
	Description string `json:"description,omitempty"`
}

// OAuth2ClientCondition contains condition information for an OAuth2Client
type OAuth2ClientCondition struct {
	Type   OAuth2ClientConditionType `json:"type"`
	Status ConditionStatus           `json:"status"`
}

type OAuth2ClientConditionType string

const (
	OAuth2ClientConditionReady = "Ready"
)

// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

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
