package models

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type MetricDTO struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMetricFromJSON(r io.Reader) (MetricDTO, error) {
	var metric MetricDTO

	if err := json.NewDecoder(r).Decode(&metric); err != nil {
		return metric, fmt.Errorf("cannot decode metric from json: %s", err)
	}

	return metric, nil
}

func NewMetricsFromJSON(r io.Reader) ([]MetricDTO, error) {
	var metrics []MetricDTO

	if err := json.NewDecoder(r).Decode(&metrics); err != nil {
		return metrics, fmt.Errorf("cannot decode metric from json: %s", err)
	}

	return metrics, nil
}

func NewGaugeMetric(id string, value float64) MetricDTO {
	mertics := MetricDTO{
		ID:    id,
		MType: GaugeType,
	}

	mertics.SetValue(value)

	return mertics
}

func NewCounterMetric(id string, delta int64) MetricDTO {
	mertics := MetricDTO{
		ID:    id,
		MType: CounterType,
	}

	mertics.SetDelta(delta)

	return mertics
}

func (m *MetricDTO) SetValue(value float64) {
	if m.Value == nil {
		m.Value = new(float64)
	}

	*m.Value = value
}

func (m *MetricDTO) SetDelta(delta int64) {
	if m.Delta == nil {
		m.Delta = new(int64)
	}

	*m.Delta = delta
}
