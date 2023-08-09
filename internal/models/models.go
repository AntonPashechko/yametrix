// Package models содержит описание структур для взаиомдействия клиента и сервера.
package models

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	GaugeType   = "gauge"   // признак метрики типа gauge
	CounterType = "counter" // признак метрики типа counter
)

// MetricDTO описывает метрику в json.
type MetricDTO struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

// NewMetricFromJSON создает экземпляр MetricDTO из json.
func NewMetricFromJSON(r io.Reader) (MetricDTO, error) {
	var metric MetricDTO

	if err := json.NewDecoder(r).Decode(&metric); err != nil {
		return metric, fmt.Errorf("cannot decode metric from json: %w", err)
	}

	return metric, nil
}

// NewMetricsFromJSON создает массив MetricDTO из json.
func NewMetricsFromJSON(r io.Reader) ([]MetricDTO, error) {
	var metrics []MetricDTO

	if err := json.NewDecoder(r).Decode(&metrics); err != nil {
		return metrics, fmt.Errorf("cannot decode metric from json: %w", err)
	}

	return metrics, nil
}

// NewGaugeMetric создает метрику типа gauge.
func NewGaugeMetric(id string, value float64) MetricDTO {
	mertics := MetricDTO{
		ID:    id,
		MType: GaugeType,
	}

	mertics.SetValue(value)

	return mertics
}

// NewCounterMetric создает метрику типа counter.
func NewCounterMetric(id string, delta int64) MetricDTO {
	mertics := MetricDTO{
		ID:    id,
		MType: CounterType,
	}

	mertics.SetDelta(delta)

	return mertics
}

// SetValue устанавливает значение gauge метрики.
func (m *MetricDTO) SetValue(value float64) {
	if m.Value == nil {
		m.Value = new(float64)
	}

	*m.Value = value
}

// SetDelta устанавливает значение counter метрики.
func (m *MetricDTO) SetDelta(delta int64) {
	if m.Delta == nil {
		m.Delta = new(int64)
	}

	*m.Delta = delta
}
