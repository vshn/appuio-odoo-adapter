package odoo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrInvalidCredentials is an error that indicates an authentication error due to missing or invalid credentials.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Session information
type Session struct {
	// SessionID is the session SessionID.
	// Is always set, no matter the authentication outcome.
	SessionID string `json:"session_id,omitempty"`
	// UID is the user's ID as an int, or the boolean `false` if authentication failed.
	UID    int `json:"uid,omitempty"`
	client *Client
}

// SearchGenericModel accepts a SearchReadModel and unmarshal the response into the given pointer.
// Depending on the JSON fields returned a custom json.Unmarshaler needs to be written since Odoo sets undefined fields to `false` instead of null.
func (c *Session) SearchGenericModel(ctx context.Context, model SearchReadModel, into interface{}) error {
	return c.executeGenericRequest(ctx, c.client.parsedURL.String()+"/web/dataset/search_read", model, into)
}

// CreateGenericModel accepts a WriteModel as a payload and executes a query to create the new data record.
func (c *Session) CreateGenericModel(ctx context.Context, model WriteModel) (int, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	resultID := 0
	err := c.executeGenericRequest(ctx, c.client.parsedURL.String()+"/web/dataset/call_kw/create", model, &resultID)
	return resultID, err
}

// UpdateGenericModel accepts a WriteModel as a payload and executes a query to update an existing data record.
func (c *Session) UpdateGenericModel(ctx context.Context, model WriteModel) (bool, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	updated := false
	err := c.executeGenericRequest(ctx, c.client.parsedURL.String()+"/web/dataset/call_kw/write", model, &updated)
	return updated, err
}

// DeleteGenericModel accepts a WriteModel as a payload and executes a query to delete an existing data record.
// For the query to succeed it is required that the Model sets an ID.
func (c *Session) DeleteGenericModel(ctx context.Context, model WriteModel) (bool, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	deleted := false
	err := c.executeGenericRequest(ctx, c.client.parsedURL.String()+"/web/dataset/call_kw/unlink", model, &deleted)
	return deleted, err
}

func (c *Session) executeGenericRequest(ctx context.Context, url string, model interface{}, into interface{}) error {
	body, err := NewJSONRPCRequest(&model).Encode()
	if err != nil {
		return newEncodingRequestError(err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return newCreatingRequestError(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+c.SessionID)

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	return c.unmarshalResponse(resp.Body, into)
}

func (c *Session) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := c.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}
	return res, nil
}

func (c *Session) unmarshalResponse(body io.ReadCloser, into interface{}) error {
	defer body.Close()
	if err := DecodeResult(body, into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
