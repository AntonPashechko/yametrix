package metrix

import (
	"net/http"
	"net/http/httptest"
	"testing"

	memstorage "github.com/AntonPashechko/yametrix/internal/storage/mem_storage"
	"github.com/stretchr/testify/assert"
)

func TestHandler_update(t *testing.T) {
	storage := memstorage.NewMemStorage()
	metrixHandler := Handler{Storage: storage}

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
			r := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			metrixHandler.update(w, r)

			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}
