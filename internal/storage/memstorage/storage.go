package memstorage

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type MemStorage struct {
	sync.Mutex

	//ЗАГЛАВНЫЕ ЧТО БЫ СРАБОТАЛ json.Marshal
	Gauge   map[string]float64
	Counter map[string]int64
}

func (m *MemStorage) clearCounter() {
	m.Counter = make(map[string]int64)
}

func NewMemStorage() storage.MetricsStorage {

	ms := &MemStorage{}
	ms.Gauge = make(map[string]float64)
	ms.Counter = make(map[string]int64)

	return ms
}

func (m *MemStorage) SetGauge(key string, value float64) {
	m.Lock()
	defer m.Unlock()

	m.Gauge[key] = value
}

func (m *MemStorage) AddCounter(key string, value int64) {
	m.Lock()
	defer m.Unlock()

	m.Counter[key] += value
}

func (m *MemStorage) GetGauge(key string) (float64, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.Gauge[key]
	return val, ok
}

func (m *MemStorage) GetCounter(key string) (int64, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.Counter[key]
	return val, ok
}

func (m *MemStorage) GetMetricsList() []string {
	m.Lock()
	defer m.Unlock()

	list := make([]string, 0, len(m.Counter)+len(m.Gauge))

	for name, value := range m.Gauge {
		strValue := utils.Float64ToStr(value)
		list = append(list, fmt.Sprintf("%s = %s", name, strValue))
	}

	for name, value := range m.Counter {
		list = append(list, fmt.Sprintf("%s = %d", name, value))
	}

	return list
}

func (m *MemStorage) GetMetrics() (map[string]float64, map[string]int64) {
	m.Lock()
	defer m.Unlock()

	defer m.clearCounter()

	return utils.DeepCopyMap(m.Gauge), utils.DeepCopyMap(m.Counter)
}

func (m *MemStorage) Marshal() ([]byte, error) {
	m.Lock()
	defer m.Unlock()

	data, err := json.Marshal(&m)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metrics: %w", err)
	}

	return data, nil
}

func (m *MemStorage) Restore(data []byte) error {
	m.Lock()
	defer m.Unlock()

	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("cannot unmarshal metrics: %w", err)
	}

	return nil
}
