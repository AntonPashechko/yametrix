package memstorage

import (
	"context"
	"testing"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/stretchr/testify/assert"
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

/*func TestMemStorage_GetGauge(t *testing.T) {

	storage := NewStorage()
	storage.SetGauge(models.NewGaugeMetric("MyGauge", 9.99))

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
}*/

/*func TestMemStorage_GetCounter(t *testing.T) {

	storage := NewStorage()
	storage.AddCounter(models.NewCounterMetric("MyCounter", 10))

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
}*/

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
				storage.SetGauge(context.TODO(), models.NewGaugeMetric(k, g))
			}
			for k, c := range tt.start.counter {
				storage.AddCounter(context.TODO(), models.NewCounterMetric(k, c))
			}

			list, err := storage.GetMetricsList(context.TODO())
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

/*func TestMemStorage_Marshal(t *testing.T) {
	tests := []struct {
		name string
		m    *MemStorage
		want string
	}{
		{
			"SimpleMarshal",
			&MemStorage{
				Gauge: map[string]models.MetricDTO{
					"MyGauge": models.MetricDTO{
						"MyGauge",
						"gauge",
						nil,
						9.99,
					},
				}, ,
				Counter: map[string]int64{
					"MyCounter": 10,
				},
			},
			`{"Gauge":{"MyGauge":9.99},"Counter":{"MyCounter":10}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.Marshal()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}*/

/*func TestMemStorage_Restore(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			"SimpleRestore",
			`{"Gauge":{"MyGauge":9.99},"Counter":{"MyCounter":10}}`,
			false,
		},
		{
			"BadDataRestore",
			`{"Gauge":{"MyGauge":9.99},"Counter":{"MyCounter":}}`,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(MemStorage)
			err := m.Restore([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}*/
