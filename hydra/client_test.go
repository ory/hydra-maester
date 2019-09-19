package hydra_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"k8s.io/utils/pointer"

	"github.com/ory/hydra-maester/hydra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	clientsEndpoint = "/clients"
	schemeHTTP      = "http"

	testID            = "test-id"
	testClient        = `{"client_id":"test-id","owner":"test-name","scope":"some,scopes","grant_types":["type1"]}`
	testClientCreated = `{"client_id":"test-id-2","client_secret":"TmGkvcY7k526","owner":"test-name-2","scope":"some,other,scopes","grant_types":["type2"]}`
	testClientUpdated = `{"client_id":"test-id-3","client_secret":"xFoPPm654por","owner":"test-name-3","scope":"yet,another,scope","grant_types":["type3"]}`
	testClientList    = `{"client_id":"test-id-4","owner":"test-name-4","scope":"scope1 scope2","grant_types":["type4"]}`
	testClientList2   = `{"client_id":"test-id-5","owner":"test-name-5","scope":"scope3 scope4","grant_types":["type5"]}`

	statusNotFoundBody            = `{"error":"Not Found","error_description":"Unable to located the requested resource","status_code":404,"request_id":"id"}`
	statusConflictBody            = `{"error":"Unable to insert or update resource because a resource with that value exists already","error_description":"","status_code":409,"request_id":"id"`
	statusInternalServerErrorBody = `the server encountered an internal error or misconfiguration and was unable to complete your request`
)

type server struct {
	statusCode int
	respBody   string
	err        error
}

var testOAuthJSONPost = &hydra.OAuth2ClientJSON{
	Scope:      "some,other,scopes",
	GrantTypes: []string{"type2"},
	Owner:      "test-name-2",
}

var testOAuthJSONPut = &hydra.OAuth2ClientJSON{
	ClientID:   pointer.StringPtr("test-id-3"),
	Scope:      "yet,another,scope",
	GrantTypes: []string{"type3"},
	Owner:      "test-name-3",
}

func TestCRUD(t *testing.T) {

	assert := assert.New(t)

	c := hydra.Client{
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

				//given
				new := tc.statusCode == http.StatusCreated

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
				o, err := c.PostOAuth2Client(testOAuthJSONPost)

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				if new {
					require.NotNil(t, o)

					assert.Equal(testOAuthJSONPost.Scope, o.Scope)
					assert.Equal(testOAuthJSONPost.GrantTypes, o.GrantTypes)
					assert.Equal(testOAuthJSONPost.Owner, o.Owner)
					assert.NotNil(o.Secret)
					assert.NotNil(o.ClientID)
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
}

func runServer(c *hydra.Client, h http.HandlerFunc) {
	s := httptest.NewServer(h)
	serverUrl, _ := url.Parse(s.URL)
	c.HydraURL = *serverUrl.ResolveReference(&url.URL{Path: clientsEndpoint})
}
