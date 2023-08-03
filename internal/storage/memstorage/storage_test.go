package memstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AntonPashechko/yametrix/internal/models"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"createMemStorage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewStorage()
			assert.NotEmpty(t, storage)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {

	storage := NewStorage()
	storage.SetGauge(context.Background(), models.NewGaugeMetric("MyGauge", 9.99))

	tests := []struct {
		name      string
		key       string
		want      float64
		wantError bool
	}{
		{"SimpleGetGauge", "MyGauge", 9.99, false},
		{"UnknownGauge", "UnGauge", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := storage.GetGauge(context.Background(), tt.key)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *val.Value)
			}
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {

	storage := NewStorage()
	storage.AddCounter(context.Background(), models.NewCounterMetric("MyCounter", 10))

	tests := []struct {
		name      string
		key       string
		want      int64
		wantError bool
	}{
		{"SimpleGetCounter", "MyCounter", 10, false},
		{"GetUnknownCounter", "UnCounter", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := storage.GetCounter(context.Background(), tt.key)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *val.Delta)
			}
		})
	}
}

func TestMemStorage_GetMetricsList(t *testing.T) {

	type metrics struct {
		gauge   map[string]float64
		counter map[string]int64
	}

	tests := []struct {
		name   string
		start  metrics
		isWant bool
		result []string
	}{
		{
			name: "SimpleGetMetricsList",
			start: metrics{
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
			name:   "EmptyMetricsList",
			isWant: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewStorage()
			for k, g := range tt.start.gauge {
				storage.SetGauge(context.Background(), models.NewGaugeMetric(k, g))
			}
			for k, c := range tt.start.counter {
				storage.AddCounter(context.Background(), models.NewCounterMetric(k, c))
			}

			list, err := storage.GetMetricsList(context.Background())
			assert.NoError(t, err)

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
