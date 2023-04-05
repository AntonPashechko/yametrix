package memstorage

import "github.com/AntonPashechko/yametrix/internal/storage"

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() storage.MertixStorage {
	ms := &MemStorage{}
	ms.gauge = make(map[string]float64)
	ms.counter = make(map[string]int64)

	return ms
}

func (m *MemStorage) SetGauge(key string, value float64) {
	m.gauge[key] = value
}

func (m *MemStorage) AddCounter(key string, value int64) {
	m.counter[key] += value
}
