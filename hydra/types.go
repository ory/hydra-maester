package hydra

// OAuth2ClientJSON represents an OAuth2 client digestible by ORY Hydra
type OAuth2ClientJSON struct {
	ClientID      *string  `json:"client_id,omitempty"`
	Name          string   `json:"client_name"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope"`
	Secret        *string  `json:"client_secret,omitempty"`
}
