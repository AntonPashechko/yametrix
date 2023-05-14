package memstorage

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type MemStorage struct {
	sync.Mutex

	//ЗАГЛАВНЫЕ ЧТО БЫ СРАБОТАЛ json.Marshal
	Gauge   map[string]models.MetricsDTO
	Counter map[string]models.MetricsDTO
}

func (m *MemStorage) clearCounter() {
	/*m.Counter = make(map[string]int64)*/
}

func NewMemStorage() storage.MetricsStorage {

	ms := &MemStorage{}
	ms.Gauge = make(map[string]models.MetricsDTO)
	ms.Counter = make(map[string]models.MetricsDTO)

	return ms
}

func (m *MemStorage) SetGauge(metric models.MetricsDTO) {
	m.Lock()
	defer m.Unlock()

	m.Gauge[metric.ID] = metric
}

func (m *MemStorage) AddCounter(metric models.MetricsDTO) models.MetricsDTO {
	m.Lock()
	defer m.Unlock()

	_, ok := m.Counter[metric.ID]
	if ok {
		*m.Counter[metric.ID].Delta += *metric.Delta
	} else {
		m.Counter[metric.ID] = metric
	}

	return m.Counter[metric.ID]
}

func (m *MemStorage) GetGauge(key string) (models.MetricsDTO, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.Gauge[key]
	return val, ok
}

func (m *MemStorage) GetCounter(key string) (models.MetricsDTO, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.Counter[key]
	return val, ok
}

func (m *MemStorage) GetMetricsList() []string {
	m.Lock()
	defer m.Unlock()

	list := make([]string, 0, len(m.Counter)+len(m.Gauge))

	for name, metric := range m.Gauge {
		strValue := utils.Float64ToStr(*metric.Value)
		list = append(list, fmt.Sprintf("%s = %s", name, strValue))
	}

	for name, metric := range m.Counter {
		list = append(list, fmt.Sprintf("%s = %d", name, *metric.Delta))
	}

	return list
}

func (m *MemStorage) GetAllMetrics() []models.MetricsDTO {
	m.Lock()
	defer m.Unlock()

	metrics := make([]models.MetricsDTO, 0)

	for _, metric := range m.Gauge {
		metrics = append(metrics, metric)
	}

	for _, metric := range m.Counter {
		metrics = append(metrics, metric)
	}

	//defer m.clearCounter()

	return metrics
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
