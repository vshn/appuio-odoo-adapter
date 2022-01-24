package odoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestClient_CreateGenericModel(t *testing.T) {
	numRequests := 0
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/create", r.RequestURI)
		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "` + uuid.NewString() + `",
			"result": 221
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	result, err := client.CreateGenericModel(newTestContext(t), &Session{}, WriteModel{})
	require.NoError(t, err)
	assert.Equal(t, 221, result)
	assert.Equal(t, 1, numRequests)
}

func TestClient_UpdateGenericModel(t *testing.T) {
	numRequests := 0
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/write", r.RequestURI)
		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "` + uuid.NewString() + `",
			"result": true
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	result, err := client.UpdateGenericModel(newTestContext(t), &Session{}, WriteModel{})
	require.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, 1, numRequests)
}

func newTestContext(t *testing.T) context.Context {
	zlogger := zaptest.NewLogger(t, zaptest.Level(zapcore.Level(-2)))
	return logr.NewContext(context.Background(), zapr.NewLogger(zlogger))
}
