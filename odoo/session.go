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
func (s *Session) SearchGenericModel(ctx context.Context, model SearchReadModel, into interface{}) error {
	return s.ExecuteQuery(ctx, "/web/dataset/search_read", model, into)
}

// CreateGenericModel accepts a WriteModel as a payload and executes a query to create the new data record.
func (s *Session) CreateGenericModel(ctx context.Context, model string, data interface{}) (int, error) {
	payload := WriteModel{
		Model:  model,
		Method: MethodCreate,
		Args:   []interface{}{data},
		KWArgs: map[string]interface{}{}, // set to non-null when serializing
	}
	resultID := 0
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/create", payload, &resultID)
	return resultID, err
}

// UpdateGenericModel accepts a WriteModel as a payload and executes a query to update an existing data record.
func (s *Session) UpdateGenericModel(ctx context.Context, model string, id int, data interface{}) error {
	if id == 0 {
		return fmt.Errorf("id cannot be zero: %v", data)
	}
	payload := WriteModel{
		Model:  model,
		Method: MethodWrite,
		Args: []interface{}{
			[]int{id},
			data,
		},
		KWArgs: map[string]interface{}{}, // set to non-null when serializing
	}
	updated := false
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/write", payload, &updated)
	return err
}

// DeleteGenericModel accepts a model identifier and data records IDs as payload and executes a query to delete multiple existing data records.
// At least one ID is required.
// It returns true if the deletion query was successful.
func (s *Session) DeleteGenericModel(ctx context.Context, model string, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("slice of ID(s) is required")
	}
	for i, id := range ids {
		if id == 0 {
			return fmt.Errorf("id cannot be zero (index: %d)", i)
		}
	}
	payload := WriteModel{
		Model:  model,
		Method: MethodDelete,
		Args:   []interface{}{ids},
		KWArgs: map[string]interface{}{}, // set to non-null when serializing
	}
	deleted := false
	err := s.ExecuteQuery(ctx, "/web/dataset/call_kw/unlink", payload, &deleted)
	return err
}

// ExecuteQuery runs a generic JSONRPC query with the given model as payload and deserializes the response.
func (s *Session) ExecuteQuery(ctx context.Context, path string, model interface{}, into interface{}) error {
	body, err := NewJSONRPCRequest(&model).Encode()
	if err != nil {
		return newEncodingRequestError(err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.parsedURL.String()+path, body)
	if err != nil {
		return newCreatingRequestError(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+s.SessionID)

	resp, err := s.sendRequest(req)
	if err != nil {
		return err
	}
	return s.unmarshalResponse(resp.Body, into)
}

func (s *Session) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := s.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}
	return res, nil
}

func (s *Session) unmarshalResponse(body io.ReadCloser, into interface{}) error {
	defer body.Close()
	if err := DecodeResult(body, into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
