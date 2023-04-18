package metrix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeMetrix(t *testing.T) {

	tests := []struct {
		name string
	}{
		{"createRuntimeMetrix"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRuntimeMetrix()
			assert.NotEmpty(t, m)
		})
	}
}

func TestRuntimeMetrix_Update(t *testing.T) {

	m := NewRuntimeMetrix()

	tests := []struct {
		name    string
		rm      *RuntimeMetrix
		wantErr bool
	}{
		{
			name:    "SimpleUpdate",
			rm:      m,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := m.Update()

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRuntimeMetrix_GetMetrix(t *testing.T) {
	m := NewRuntimeMetrix()

	tests := []struct {
		name string
		rm   *RuntimeMetrix
	}{
		{
			name: "SimpleGetMetrix",
			rm:   m,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := m.Update()
			if assert.Nil(t, err) {
				g, c := m.GetMetrix()
				assert.NotEmpty(t, g, "empty gauge")
				assert.NotEmpty(t, c, "empty counter")
			}
		})
	}
}
