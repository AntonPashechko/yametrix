package memstorage

import (
	"fmt"
	"sync"

	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type MemStorage struct {
	sync.Mutex

	gauge   map[string]float64
	counter map[string]int64
}

func (m *MemStorage) clearCounter() {
	m.counter = make(map[string]int64)
}

func NewMemStorage() storage.MetrixStorage {

	ms := &MemStorage{}
	ms.gauge = make(map[string]float64)
	ms.counter = make(map[string]int64)

	return ms
}

func (m *MemStorage) SetGauge(key string, value float64) {
	m.Lock()
	defer m.Unlock()

	m.gauge[key] = value
}

func (m *MemStorage) AddCounter(key string, value int64) {
	m.Lock()
	defer m.Unlock()

	m.counter[key] += value
}

func (m *MemStorage) GetGauge(key string) (float64, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.gauge[key]
	return val, ok
}

func (m *MemStorage) GetCounter(key string) (int64, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.counter[key]
	return val, ok
}

func (m *MemStorage) GetMetrixList() []string {
	m.Lock()
	defer m.Unlock()

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

func (m *MemStorage) GetMetrix() (map[string]float64, map[string]int64) {
	m.Lock()
	defer m.Unlock()

	defer m.clearCounter()

	return utils.DeepCopyMap(m.gauge), utils.DeepCopyMap(m.counter)
}
