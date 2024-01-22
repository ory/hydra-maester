// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	"github.com/ory/hydra-maester/hydra"
)

const (
	clientsEndpoint = "/clients"
	schemeHTTP      = "http"

	testID                        = "test-id"
	testClient                    = `{"client_id":"test-id","owner":"test-name","scope":"some,scopes","grant_types":["type1"],"token_endpoint_auth_method":"client_secret_basic"}`
	testClientCreated             = `{"client_id":"test-id-2","client_secret":"TmGkvcY7k526","owner":"test-name-2","scope":"some,other,scopes","grant_types":["type2"],"audience":["audience-a","audience-b"],"token_endpoint_auth_method":"client_secret_basic","backchannel_logout_uri":"https://localhost/backchannel-logout","frontchannel_logout_uri":"https://localhost/frontchannel-logout"}`
	testClientUpdated             = `{"client_id":"test-id-3","client_secret":"xFoPPm654por","owner":"test-name-3","scope":"yet,another,scope","grant_types":["type3"],"audience":["audience-c"],"token_endpoint_auth_method":"client_secret_basic"}`
	testClientList                = `{"client_id":"test-id-4","owner":"test-name-4","scope":"scope1 scope2","grant_types":["type4"],"token_endpoint_auth_method":"client_secret_basic"}`
	testClientList2               = `{"client_id":"test-id-5","owner":"test-name-5","scope":"scope3 scope4","grant_types":["type5"],"token_endpoint_auth_method":"client_secret_basic"}`
	testClientWithMetadataCreated = `{"client_id":"test-id-21","client_secret":"TmGkvcY7k526","owner":"test-name-21","scope":"some,other,scopes","grant_types":["type2"],"token_endpoint_auth_method":"client_secret_basic","metadata":{"property1":1,"property2":"2"},"backchannel_logout_uri":"https://localhost/backchannel-logout","frontchannel_logout_uri":"https://localhost/frontchannel-logout"}`

	statusNotFoundBody            = `{"error":"Not Found","error_description":"Unable to locate the requested resource","status_code":404,"request_id":"id"}`
	statusUnauthorizedBody        = `{"error":"The request could not be authorized","error_description":"The requested OAuth 2.0 client does not exist or you did not provide the necessary credentials","status_code":401,"request_id":"id"}`
	statusConflictBody            = `{"error":"Unable to insert or update resource because a resource with that value exists already","error_description":"","status_code":409,"request_id":"id"`
	statusInternalServerErrorBody = "the server encountered an internal error or misconfiguration and was unable to complete your request"
)

type server struct {
	statusCode int
	respBody   string
	err        error
}

var testOAuthJSONPost = &hydra.OAuth2ClientJSON{
	Scope:                             "some,other,scopes",
	GrantTypes:                        []string{"type2"},
	Owner:                             "test-name-2",
	Audience:                          []string{"audience-a", "audience-b"},
	FrontChannelLogoutURI:             "https://localhost/frontchannel-logout",
	FrontChannelLogoutSessionRequired: false,
	BackChannelLogoutURI:              "https://localhost/backchannel-logout",
	BackChannelLogoutSessionRequired:  false,
}

var testOAuthJSONPut = &hydra.OAuth2ClientJSON{
	ClientID:   ptr.To("test-id-3"),
	Scope:      "yet,another,scope",
	GrantTypes: []string{"type3"},
	Owner:      "test-name-3",
	Audience:   []string{"audience-c"},
}

func TestCRUD(t *testing.T) {

	assert := assert.New(t)

	c := hydra.InternalClient{
		HTTPClient: &http.Client{},
		HydraURL:   url.URL{Scheme: schemeHTTP},
	}

	t.Run("method=get", func(t *testing.T) {

		for d, tc := range map[string]server{
			"getting registered client": {
				http.StatusOK,
				testClient,
				nil,
			},
			"getting unregistered client": {
				http.StatusNotFound,
				statusNotFoundBody,
				nil,
			},
			"getting unauthorized request": {
				http.StatusUnauthorized,
				statusUnauthorizedBody,
				nil,
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {

				//given
				shouldFind := tc.statusCode == http.StatusOK

				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(fmt.Sprintf("%s/%s", c.HydraURL.String(), testID), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodGet, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					if shouldFind {
						w.Header().Set("Content-type", "application/json")
					}
				})
				runServer(&c, h)

				//when
				o, found, err := c.GetOAuth2Client(testID)

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				assert.Equal(shouldFind, found)
				if shouldFind {
					require.NotNil(t, o)
					var expected hydra.OAuth2ClientJSON
					json.Unmarshal([]byte(testClient), &expected)
					assert.Equal(&expected, o)
				}
			})
		}
	})

	t.Run("method=post", func(t *testing.T) {

		for d, tc := range map[string]server{
			"with new client": {
				http.StatusCreated,
				testClientCreated,
				nil,
			},
			"with new client with metadata": {
				http.StatusCreated,
				testClientWithMetadataCreated,
				nil,
			},
			"with existing client": {
				http.StatusConflict,
				statusConflictBody,
				errors.New("requested ID already exists"),
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {
				var (
					err      error
					o        *hydra.OAuth2ClientJSON
					expected *hydra.OAuth2ClientJSON
				)
				//given
				new := tc.statusCode == http.StatusCreated
				newWithMetadata := d == "with new client with metadata"

				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(c.HydraURL.String(), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodPost, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					if new {
						w.Header().Set("Content-type", "application/json")
					}
				})
				runServer(&c, h)

				//when
				if newWithMetadata {
					meta, _ := json.Marshal(map[string]interface{}{
						"property1": float64(1),
						"property2": "2",
					})
					var testOAuthJSONPost2 = &hydra.OAuth2ClientJSON{
						Scope:                             "some,other,scopes",
						GrantTypes:                        []string{"type2"},
						Owner:                             "test-name-21",
						Metadata:                          meta,
						FrontChannelLogoutURI:             "https://localhost/frontchannel-logout",
						FrontChannelLogoutSessionRequired: false,
						BackChannelLogoutURI:              "https://localhost/backchannel-logout",
						BackChannelLogoutSessionRequired:  false,
					}
					o, err = c.PostOAuth2Client(testOAuthJSONPost2)
					expected = testOAuthJSONPost2
				} else {
					o, err = c.PostOAuth2Client(testOAuthJSONPost)
					expected = testOAuthJSONPost
				}

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				if new {
					require.NotNil(t, o)
					assert.Equal(expected.Scope, o.Scope)
					assert.Equal(expected.GrantTypes, o.GrantTypes)
					assert.Equal(expected.Owner, o.Owner)
					assert.Equal(expected.Audience, o.Audience)
					assert.NotNil(o.Secret)
					assert.NotNil(o.ClientID)
					assert.NotNil(o.TokenEndpointAuthMethod)
					assert.Equal(expected.FrontChannelLogoutURI, o.FrontChannelLogoutURI)
					assert.Equal(expected.FrontChannelLogoutSessionRequired, o.FrontChannelLogoutSessionRequired)
					assert.Equal(expected.BackChannelLogoutURI, o.BackChannelLogoutURI)
					assert.Equal(expected.BackChannelLogoutSessionRequired, o.BackChannelLogoutSessionRequired)
					if expected.TokenEndpointAuthMethod != "" {
						assert.Equal(expected.TokenEndpointAuthMethod, o.TokenEndpointAuthMethod)
					}
					if newWithMetadata {
						assert.NotNil(o.Metadata)
						assert.True(len(o.Metadata) > 0)
						for key := range o.Metadata {
							assert.Equal(o.Metadata[key], expected.Metadata[key])
						}
					} else {
						assert.Nil(o.Metadata)
					}
				}
			})
		}
	})

	t.Run("method=put", func(t *testing.T) {
		for d, tc := range map[string]server{
			"with registered client": {
				http.StatusOK,
				testClientUpdated,
				nil,
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {

				ok := tc.statusCode == http.StatusOK

				//given
				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(fmt.Sprintf("%s/%s", c.HydraURL.String(), *testOAuthJSONPut.ClientID), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodPut, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					if ok {
						w.Header().Set("Content-type", "application/json")
					}
				})
				runServer(&c, h)

				//when
				o, err := c.PutOAuth2Client(testOAuthJSONPut)

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				if ok {
					require.NotNil(t, o)

					assert.Equal(testOAuthJSONPut.Scope, o.Scope)
					assert.Equal(testOAuthJSONPut.GrantTypes, o.GrantTypes)
					assert.Equal(testOAuthJSONPut.ClientID, o.ClientID)
					assert.Equal(testOAuthJSONPut.Owner, o.Owner)
					assert.Equal(testOAuthJSONPut.Audience, o.Audience)
					assert.NotNil(o.Secret)
				}
			})
		}
	})

	t.Run("method=delete", func(t *testing.T) {

		for d, tc := range map[string]server{
			"with registered client": {
				statusCode: http.StatusNoContent,
			},
			"with unregistered client": {
				statusCode: http.StatusNotFound,
				respBody:   statusNotFoundBody,
			},
			"internal server error when requesting": {
				statusCode: http.StatusInternalServerError,
				respBody:   statusInternalServerErrorBody,
				err:        errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {

				//given
				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(fmt.Sprintf("%s/%s", c.HydraURL.String(), testID), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodDelete, req.Method)
					w.WriteHeader(tc.statusCode)
				})
				runServer(&c, h)

				//when
				err := c.DeleteOAuth2Client(testID)

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}
			})
		}
	})

	t.Run("method=list", func(t *testing.T) {

		for d, tc := range map[string]server{
			"no clients": {
				http.StatusOK,
				`[]`,
				nil,
			},
			"one client": {
				http.StatusOK,
				fmt.Sprintf("[%s]", testClientList),
				nil,
			},
			"more clients": {
				http.StatusOK,
				fmt.Sprintf("[%s,%s]", testClientList, testClientList2),
				nil,
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {

				//given
				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(c.HydraURL.String(), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodGet, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					w.Header().Set("Content-type", "application/json")

				})
				runServer(&c, h)

				//when
				list, err := c.ListOAuth2Client()

				//then
				if tc.err == nil {
					require.NoError(t, err)
					require.NotNil(t, list)
					var expectedList []*hydra.OAuth2ClientJSON
					json.Unmarshal([]byte(tc.respBody), &expectedList)
					assert.Equal(expectedList, list)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}
			})
		}
	})

	t.Run("default parameters", func(t *testing.T) {
		var input = &hydra.OAuth2ClientJSON{
			Scope:      "some,other,scopes",
			GrantTypes: []string{"type2"},
			Owner:      "test-name-2",
		}
		assert.Equal(input.TokenEndpointAuthMethod, "")
		b, _ := json.Marshal(input)
		payload := string(b)
		assert.Equal(strings.Index(payload, "token_endpoint_auth_method"), -1)

		input = &hydra.OAuth2ClientJSON{
			Scope:                   "some,other,scopes",
			GrantTypes:              []string{"type2"},
			Owner:                   "test-name-3",
			TokenEndpointAuthMethod: "none",
		}
		b, _ = json.Marshal(input)
		payload = string(b)
		assert.True(strings.Index(payload, "token_endpoint_auth_method") > 0)
	})
}

func runServer(c *hydra.InternalClient, h http.HandlerFunc) {
	s := httptest.NewServer(h)
	serverUrl, _ := url.Parse(s.URL)
	c.HydraURL = *serverUrl.ResolveReference(&url.URL{Path: clientsEndpoint})
}
