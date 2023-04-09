package metrix

import (
	"net/http"
	"net/http/httptest"
	"testing"

	memstorage "github.com/AntonPashechko/yametrix/internal/storage/mem_storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	metrixHandler := NewMetrixHandler(storage)
	metrixHandler.Register(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{"/update/gauge/testGauge/100", "/update/gauge/testGauge/100", http.StatusOK},
		{"/update/gauge/", "/update/gauge/", http.StatusNotFound},
		{"/update/gauge/testGauge/none", "/update/gauge/testGauge/none", http.StatusBadRequest},

		{"/update/counter/testCounter/100", "/update/counter/testCounter/100", http.StatusOK},
		{"/update/counter/", "/update/counter/", http.StatusNotFound},
		{"/update/counter/testCounter/none", "/update/counter/testCounter/none", http.StatusBadRequest},

		{"/update/unknown/testCounter/100", "/update/unknown/testCounter/100", http.StatusNotImplemented},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, "POST", tt.url)
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			resp.Body.Close()
		})
	}
}
