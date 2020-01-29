package hydra

import "k8s.io/utils/pointer"

// OAuth2ClientJSON represents an OAuth2 client digestible by ORY Hydra
type OAuth2ClientJSON struct {
	ClientID                *string  `json:"client_id,omitempty"`
	Secret                  *string  `json:"client_secret,omitempty"`
	GrantTypes              []string `json:"grant_types"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	Audience                []string `json:"audience,omitempty"`
	Scope                   string   `json:"scope"`
	Owner                   string   `json:"owner"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
}

// Oauth2ClientCredentials represents client ID and password fetched from a Kubernetes secret
type Oauth2ClientCredentials struct {
	ID       []byte
	Password []byte
}

func (oj *OAuth2ClientJSON) WithCredentials(credentials *Oauth2ClientCredentials) *OAuth2ClientJSON {
	oj.ClientID = pointer.StringPtr(string(credentials.ID))
	oj.Secret = pointer.StringPtr(string(credentials.Password))
	return oj
}
