package odoo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
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

type debugTransport struct {
	pwRe *regexp.Regexp
}

func newDebugTransport() *debugTransport {
	return &debugTransport{
		pwRe: regexp.MustCompile(`("password":\s?").+("[,}])`),
	}
}

// UseDebugLogger sets the http.Transport field of the internal http client with a transport implementation that logs the raw contents of requests and responses.
// The logger is retrieved from the request's context via logr.FromContextOrDiscard.
// The log level used is '2'.
// Any "password":"..." byte content is replaced with a placeholder to avoid leaking credentials.
// Still, this should not be called in production as other sensitive information might be leaked.
// This method is meant to be called before any requests are made (for example after setting up the Client).
func (c Client) UseDebugLogger(enabled bool) {
	if enabled {
		c.http.Transport = newDebugTransport()
	}
}

// RoundTrip implements http.RoundTripper.
func (t *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	logger := logr.FromContextOrDiscard(r.Context())
	if r.Body != nil {
		reqBody, _ := r.GetBody()
		defer reqBody.Close()
		buf, _ := ioutil.ReadAll(reqBody)
		buf = t.pwRe.ReplaceAll(buf, []byte(`$1[confidential]$2`))
		logger.V(2).Info(fmt.Sprintf("%s %s ---> %s", r.Method, r.URL.Path, string(buf)))
	}

	res, err := http.DefaultTransport.RoundTrip(r)

	if res.Body != nil {
		defer res.Body.Close()
		buf, _ := ioutil.ReadAll(res.Body)
		logger.V(2).Info(fmt.Sprintf("%s %s <--- %s", r.Method, r.URL.Path, string(buf)))
		res.Body = io.NopCloser(bytes.NewReader(buf))
	}

	return res, err
}

func (c *Client) makeRequest(ctx context.Context, session *Session, body io.Reader) (*http.Response, error) {
	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/web/dataset/search_read", body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+session.SessionID)

	// Send request
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
	if err := DecodeResult(body, &into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
