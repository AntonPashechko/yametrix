package models

type MetricsDTO struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMetricsDTO(id string, mType string, delta *int64, value *float64) MetricsDTO {
	mertics := MetricsDTO{
		ID:    id,
		MType: mType,
	}

	if delta != nil {
		mertics.SetDelta(*delta)
	} else if value != nil {
		mertics.SetValue(*value)
	}

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
