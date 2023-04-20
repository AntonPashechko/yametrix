package sender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPSendWorker(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"createHTTPSendWorker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := NewHTTPSendWorker(nil, "")
			assert.NotEmpty(t, sch)
		})
	}
}

func Test_httpSendWorker_Work(t *testing.T) {
	tests := []struct {
		name    string
		m       *httpSendWorker
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Work(); (err != nil) != tt.wantErr {
				t.Errorf("httpSendWorker.Work() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
