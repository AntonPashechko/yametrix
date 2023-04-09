package memstorage

import (
	"fmt"

	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

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

func (m *MemStorage) GetGauge(key string) (float64, bool) {
	val, ok := m.gauge[key]
	return val, ok
}

func (m *MemStorage) GetCounter(key string) (int64, bool) {
	val, ok := m.counter[key]
	return val, ok
}

func (m *MemStorage) GetMetrixList() []string {
	list := make([]string, 0, len(m.counter)+len(m.gauge))

	for name, value := range m.gauge {
		strValue := utils.Float64ToStr(value)
		list = append(list, fmt.Sprintf("%s = %s", name, strValue))
	}

	for name, value := range m.counter {
		list = append(list, fmt.Sprintf("%s = %d", name, value))
	}

	return list
}
