package odoo

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestClient_CreateGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/create", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"create",
				"args":[					
					"data"
				],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": 221
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	m := NewCreateModel("model", "data")
	result, err := client.CreateGenericModel(newTestContext(t), &Session{}, m)
	require.NoError(t, err)
	assert.Equal(t, 221, result)
	assert.Equal(t, 1, numRequests)
}

func TestClient_UpdateGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/write", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"write",
				"args":[
					[1],
					"data"
				],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": true
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	m, err := NewUpdateModel("model", 1, "data")
	require.NoError(t, err)
	result, err := client.UpdateGenericModel(newTestContext(t), &Session{}, m)
	require.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, 1, numRequests)
}

func TestClient_DeleteGenericModel(t *testing.T) {
	numRequests := 0
	uuidGenerator = func() string {
		return "fakeID"
	}
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/dataset/call_kw/unlink", r.RequestURI)

		buf, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"id":"fakeID",
			"jsonrpc":"2.0",
			"method":"call",
			"params":{
				"model":"model",
				"method":"unlink",
				"args":[[100]],
				"kwargs":{}
			}}`, string(buf))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "fakeID",
			"result": true
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	m, err := NewDeleteModel("model", []int{100})
	require.NoError(t, err)
	result, err := client.DeleteGenericModel(newTestContext(t), &Session{}, m)
	require.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, 1, numRequests)
}

func newTestContext(t *testing.T) context.Context {
	zlogger := zaptest.NewLogger(t, zaptest.Level(zapcore.Level(-2)))
	return logr.NewContext(context.Background(), zapr.NewLogger(zlogger))
}
