package odoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is the base struct that holds information required to talk to Odoo
type Client struct {
	parsedURL *url.URL
	db        string
	http      *http.Client
}

// NewClient returns a new client with its basic fields set.
// It returns error if baseURL is not parseable with url.Parse.
func NewClient(baseURL, db string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("proper URL format is required: %w", err)
	}
	return &Client{
		parsedURL: u,
		db:        db,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     nil, // don't save any cookies!
		}}, nil
}

type loginParams struct {
	DB       string `json:"db,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

// Login tries to authenticate the user against Odoo.
// It returns a session if authentication was successful. An error is returned if
//  - the credentials were wrong,
//  - encoding or sending the request,
//  - or decoding the request failed.
func (c Client) Login(ctx context.Context, login, password string) (*Session, error) {
	resp, err := c.requestSession(ctx, login, password)
	if err != nil {
		return nil, err
	}

	return c.decodeSession(resp)
}

func (c Client) requestSession(ctx context.Context, login string, password string) (*http.Response, error) {
	// Prepare request
	body, err := NewJSONRPCRequest(loginParams{c.db, login, password}).Encode()
	if err != nil {
		return nil, newEncodingRequestError(err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.parsedURL.String()+"/web/session/authenticate", body)
	if err != nil {
		return nil, newCreatingRequestError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login: sending HTTP request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login: expected HTTP status 200 OK, got %s", resp.Status)
	}
	return resp, nil
}

func (c *Client) decodeSession(res *http.Response) (*Session, error) {
	// Decode response
	// We don't use DecodeResult here because we're interested in whether unmarshalling the result failed.
	// If so, this is likely because "uid" is set to `false` which indicates an authentication failure.
	var response JSONRPCResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("login: decode response: %w", err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("error from Odoo: %v", response.Error)
	}

	// Decode session
	var session Session
	if err := json.Unmarshal(*response.Result, &session); err != nil {
		// UID is not set, authentication failed
		return nil, ErrInvalidCredentials
	}
	session.client = c
	return &session, nil
}
