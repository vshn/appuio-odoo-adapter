package odoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the base struct that holds information required to talk to Odoo
type Client struct {
	baseURL   string
	parsedURL *url.URL
	db        string
	http      *http.Client
}

// NewClient returns a new client with its basic fields set.
// It panics if baseURL is not parseable with url.Parse.
func NewClient(baseURL, db string) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Errorf("proper URL format is required: %w", err))
	}
	return &Client{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		parsedURL: u,
		db:        db,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     nil, // don't save any cookies!
		}}
}

// SearchGenericModel accepts a SearchReadModel and unmarshal the response into the given pointer.
// Depending on the JSON fields returned a custom json.Unmarshaler needs to be written since Odoo sets undefined fields to `false` instead of null.
func (c Client) SearchGenericModel(ctx context.Context, session *Session, model SearchReadModel, into interface{}) error {
	return c.executeGenericRequest(ctx, session, c.parsedURL.String()+"/web/dataset/search_read", model, into)
}

// CreateGenericModel accepts a WriteModel as a payload and executes a query to create the new data record.
func (c Client) CreateGenericModel(ctx context.Context, session *Session, model WriteModel) (int, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	resultID := 0
	err := c.executeGenericRequest(ctx, session, c.parsedURL.String()+"/web/dataset/call_kw/create", model, &resultID)
	return resultID, err
}

// UpdateGenericModel accepts a WriteModel as a payload and executes a query to update an existing data record.
func (c Client) UpdateGenericModel(ctx context.Context, session *Session, model WriteModel) (bool, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	updated := false
	err := c.executeGenericRequest(ctx, session, c.parsedURL.String()+"/web/dataset/call_kw/write", model, &updated)
	return updated, err
}

// DeleteGenericModel accepts a WriteModel as a payload and executes a query to delete an existing data record.
// For the query to succeed it is required that the Model sets an ID.
func (c Client) DeleteGenericModel(ctx context.Context, session *Session, model WriteModel) (bool, error) {
	if model.KWArgs == nil {
		model.KWArgs = map[string]interface{}{} // set to non-null when serializing
	}
	deleted := false
	err := c.executeGenericRequest(ctx, session, c.parsedURL.String()+"/web/dataset/call_kw/unlink", model, &deleted)
	return deleted, err
}

func (c Client) executeGenericRequest(ctx context.Context, session *Session, url string, model interface{}, into interface{}) error {
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
	req.Header.Set("cookie", "session_id="+session.SessionID)

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	return c.unmarshalResponse(resp.Body, into)
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}
	return res, nil
}

func (c *Client) unmarshalResponse(body io.ReadCloser, into interface{}) error {
	defer body.Close()
	if err := DecodeResult(body, into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
