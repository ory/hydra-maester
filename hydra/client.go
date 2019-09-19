package hydra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type Client struct {
	HydraURL   url.URL
	HTTPClient *http.Client
}

func (c *Client) GetOAuth2Client(id string) (*OAuth2ClientJSON, bool, error) {

	var jsonClient *OAuth2ClientJSON

	req, err := c.newRequest(http.MethodGet, id, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := c.do(req, &jsonClient)
	if err != nil {
		return nil, false, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return jsonClient, true, nil
	case http.StatusNotFound:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}

func (c *Client) ListOAuth2Client() ([]*OAuth2ClientJSON, error) {

	var jsonClientList []*OAuth2ClientJSON

	req, err := c.newRequest(http.MethodGet, "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, &jsonClientList)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return jsonClientList, nil
	default:
		return nil, fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}

func (c *Client) PostOAuth2Client(o *OAuth2ClientJSON) (*OAuth2ClientJSON, error) {

	var jsonClient *OAuth2ClientJSON

	req, err := c.newRequest(http.MethodPost, "", o)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, &jsonClient)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return jsonClient, nil
	case http.StatusConflict:
		return nil, fmt.Errorf("%s %s http request failed: requested ID already exists", req.Method, req.URL)
	default:
		return nil, fmt.Errorf("%s %s http request returned unexpected status code: %s", req.Method, req.URL, resp.Status)
	}
}

func (c *Client) PutOAuth2Client(o *OAuth2ClientJSON) (*OAuth2ClientJSON, error) {

	var jsonClient *OAuth2ClientJSON

	req, err := c.newRequest(http.MethodPut, *o.ClientID, o)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, &jsonClient)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s http request returned unexpected status code: %s", req.Method, req.URL, resp.Status)
	}

	return jsonClient, nil
}

func (c *Client) DeleteOAuth2Client(id string) error {

	req, err := c.newRequest(http.MethodDelete, id, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		fmt.Printf("client with id %s does not exist", id)
		return nil
	default:
		return fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}

func (c *Client) newRequest(method, relativePath string, body interface{}) (*http.Request, error) {

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	u := c.HydraURL
	u.Path = path.Join(u.Path, relativePath)

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil

}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if v != nil && resp.StatusCode < 300 {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
