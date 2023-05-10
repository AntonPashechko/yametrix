package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	memstorage "github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequestWithBody(t *testing.T, ts *httptest.Server, method, path, body string) *resty.Response {

	req := resty.New().R()
	req.Method = method
	req.URL = ts.URL + path

	if len(body) > 0 {
		req.SetHeader("Content-Type", "application/json")
		req.SetBody(body)
	}

	resp, err := req.Send()
	require.NoError(t, err)
	return resp
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp
}

func TestHandler_update(t *testing.T) {
	storage := memstorage.NewMemStorage()
	router := chi.NewRouter()
	metrixHandler := NewMetrixHandler(storage, nil)
	metrixHandler.Register(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	tests := []struct {
		url          string
		expectedCode int
	}{
		{"/update/gauge/testGauge/100", http.StatusOK},
		{"/update/gauge/", http.StatusNotFound},
		{"/update/gauge/testGauge/none", http.StatusBadRequest},

		{"/update/counter/testCounter/100", http.StatusOK},
		{"/update/counter/", http.StatusNotFound},
		{"/update/counter/testCounter/none", http.StatusBadRequest},

		{"/update/unknown/testCounter/100", http.StatusNotImplemented},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp := testRequest(t, ts, "POST", tt.url)
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			resp.Body.Close()
		})
	}
}

func TestHandler_updateJson(t *testing.T) {
	storage := memstorage.NewMemStorage()
	router := chi.NewRouter()
	metrixHandler := NewMetrixHandler(storage, nil)
	metrixHandler.Register(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	tests := []struct {
		name         string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Simple test gauge",
			body:         `{"id":"test_gauge","type":"gauge","value":123.123}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"test_gauge","type":"gauge","value":123.123}`,
		},
		{
			name:         "Simple test counter",
			body:         `{"id":"test_counter","type":"counter","delta":2}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"test_counter","type":"counter","delta":2}`,
		},
		{
			name:         "Simple add counter",
			body:         `{"id":"test_counter","type":"counter","delta":2}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"test_counter","type":"counter","delta":4}`,
		},
		{
			name:         "Empty body",
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequestWithBody(t, ts, "POST", "/update/", tt.body)
			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			replacer := strings.NewReplacer("\r", "", "\n", "")

			assert.Equal(t, tt.expectedBody, replacer.Replace(string(resp.Body())), "Значение ответа не совпадает с ожидаемым")
		})
	}
}
