package memstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"createMemStorage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			assert.NotEmpty(t, storage)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {

	storage := NewMemStorage()
	storage.SetGauge("MyGauge", 9.99)

	tests := []struct {
		name  string
		key   string
		want  float64
		want1 bool
	}{
		{"SimpleGetGauge", "MyGauge", 9.99, true},
		{"UnknownGauge", "UnGauge", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := storage.GetGauge(tt.key)
			assert.Equal(t, tt.want, val)
			assert.Equal(t, tt.want1, ok)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {

	storage := NewMemStorage()
	storage.AddCounter("MyCounter", 10)

	tests := []struct {
		name  string
		key   string
		want  int64
		want1 bool
	}{
		{"SimpleAddCounter", "MyCounter", 10, true},
		{"UnknownCounter", "UnCounter", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := storage.GetCounter(tt.key)
			assert.Equal(t, tt.want, val)
			assert.Equal(t, tt.want1, ok)
		})
	}
}

func TestMemStorage_GetMetrixList(t *testing.T) {

	type metrix struct {
		gauge   map[string]float64
		counter map[string]int64
	}

	tests := []struct {
		name   string
		start  metrix
		isWant bool
		result []string
	}{
		{
			name: "SimpleGetMetrixList",
			start: metrix{
				gauge: map[string]float64{
					"MyGauge": 9.99,
				},
				counter: map[string]int64{
					"MyCounter": 10,
				},
			},
			isWant: true,
			result: []string{
				"MyGauge = 9.99",
				"MyCounter = 10",
			},
		},
		{
			name:   "EmptyMetrixList",
			isWant: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			for k, g := range tt.start.gauge {
				storage.SetGauge(k, g)
			}
			for k, c := range tt.start.counter {
				storage.AddCounter(k, c)
			}

			list := storage.GetMetrixList()

			if tt.isWant {
				if assert.NotEmpty(t, list) {
					assert.ElementsMatch(t, list, tt.result)
				}
			} else {
				assert.Empty(t, list)
			}
		})
	}
}
