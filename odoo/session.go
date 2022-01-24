package odoo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrInvalidCredentials is an error that indicates an authentication error due to missing or invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Session information
type Session struct {
	// ID is the session ID.
	// Is always set, no matter the authentication outcome.
	ID string `json:"session_id,omitempty"`
	// UID is the user's ID as an int, or the boolean `false` if authentication failed.
	UID int `json:"uid,omitempty"`
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
		return nil, fmt.Errorf("encoding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.parsedURL.String()+"/web/session/authenticate", body)
	if err != nil {
		return nil, fmt.Errorf("login: building HTTP request: %w", err)
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

func (c Client) decodeSession(res *http.Response) (*Session, error) {
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

	return &session, nil
}
