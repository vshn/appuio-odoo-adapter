package odoo

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_Success(t *testing.T) {
	var (
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testUID      = rand.Int()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/session/authenticate", r.RequestURI)

		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		body := string(b)
		assert.Contains(t, body, `"db":"TestDB"`)
		assert.Contains(t, body, `"login":"`+testLogin+`"`)
		assert.Contains(t, body, `"password":"`+testPassword+`"`)

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": 1,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": ` + strconv.Itoa(testUID) + `,
				"user_context": {
					"lang": "en_US",
					"tz": "Europe/Zurich",
					"uid": ` + strconv.Itoa(testUID) + `
				},
				"username": "` + testLogin + `"
			}
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Login
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	session, err := client.Login(newTestContext(t), testLogin, testPassword)
	require.NoError(t, err)
	assert.Equal(t, testUID, session.UID)
	assert.Equal(t, testSID, session.SessionID)
	assert.Equal(t, 1, numRequests)
}

func TestLogin_BadCredentials(t *testing.T) {
	var (
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/session/authenticate", r.RequestURI)
		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": null,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": false,
				"user_context": {},
				"username": "` + testLogin + `"
			}
		}`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	session, err := client.Login(newTestContext(t), testLogin, testPassword)
	require.EqualError(t, err, "invalid credentials")
	assert.Nil(t, session)
	assert.Equal(t, 1, numRequests)
}

func TestLogin_BadResponse(t *testing.T) {
	numRequests := 0
	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "xxx",
			"error": {
			  "message": "Odoo Server Error",
			  "code": 200,
			  "data": {
				"debug": "Traceback xxx",
				"message": "",
				"name": "werkzeug.exceptions.Foo",
				"arguments": []
			  }
			}
		  }`))
		require.NoError(t, err)
	}))
	defer odooMock.Close()

	// Do request
	client := NewClient(odooMock.URL, "TestDB")
	client.UseDebugLogger(true)
	session, err := client.Login(newTestContext(t), "", "")
	require.EqualError(t, err, "error from Odoo: &{Odoo Server Error 200 map[arguments:[] debug:Traceback xxx message: name:werkzeug.exceptions.Foo]}")
	assert.Nil(t, session)
	assert.Equal(t, 1, numRequests)
}
