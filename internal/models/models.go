package models

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type MetricsDTO struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewGaugeMetric(id string, value float64) MetricsDTO {
	mertics := MetricsDTO{
		ID:    id,
		MType: GaugeType,
	}

	mertics.SetValue(value)

	return mertics
}

func NewCounterMetric(id string, delta int64) MetricsDTO {
	mertics := MetricsDTO{
		ID:    id,
		MType: CounterType,
	}

	mertics.SetDelta(delta)

	return mertics
}

func (m *MetricsDTO) SetValue(value float64) {
	if m.Value == nil {
		m.Value = new(float64)
	}

	*m.Value = value
}

func (m *MetricsDTO) SetDelta(delta int64) {
	if m.Delta == nil {
		m.Delta = new(int64)
	}

	*m.Delta = delta
}
