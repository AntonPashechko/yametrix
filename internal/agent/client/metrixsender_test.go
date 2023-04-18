package client

import (
	"net/http/httptest"
	"testing"

	"github.com/AntonPashechko/yametrix/internal/handlers/metrix"
	memstorage "github.com/AntonPashechko/yametrix/internal/storage/mem_storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrixClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"createMetrixClient"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := NewMetrixClient(nil, "")
			assert.NotEmpty(t, sch)
		})
	}
}

func TestMetrixHTTPClient_Send(t *testing.T) {
	storage := memstorage.NewMemStorage()
	router := chi.NewRouter()
	metrixHandler := metrix.NewMetrixHandler(storage)
	metrixHandler.Register(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	tests := []struct {
		name    string
		mhc     *MetrixHTTPClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mhc.Send(); (err != nil) != tt.wantErr {
				t.Errorf("MetrixHTTPClient.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
