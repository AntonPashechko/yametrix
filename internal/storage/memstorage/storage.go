package memstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

var mux sync.Mutex

type Storage struct {
	//ЗАГЛАВНЫЕ ЧТО БЫ СРАБОТАЛ json.Marshal
	Gauge   map[string]models.MetricDTO
	Counter map[string]models.MetricDTO
}

func (m *Storage) clearCounter() {
	m.Counter = make(map[string]models.MetricDTO)
}

func NewStorage() *Storage {

	ms := &Storage{}
	ms.Gauge = make(map[string]models.MetricDTO)
	ms.Counter = make(map[string]models.MetricDTO)

	return ms
}

func (m *Storage) SetGauge(ctx context.Context, metric models.MetricDTO) error {
	mux.Lock()
	defer mux.Unlock()

	m.Gauge[metric.ID] = metric
	return nil
}

func (m *Storage) AddCounter(ctx context.Context, metric models.MetricDTO) (*models.MetricDTO, error) {
	mux.Lock()
	defer mux.Unlock()

	_, ok := m.Counter[metric.ID]
	if ok {
		*m.Counter[metric.ID].Delta += *metric.Delta
	} else {
		m.Counter[metric.ID] = metric
	}

	val := m.Counter[metric.ID]

	return &val, nil
}

func (m *Storage) GetGauge(ctx context.Context, key string) (*models.MetricDTO, error) {
	mux.Lock()
	defer mux.Unlock()

	val, ok := m.Gauge[key]
	if !ok {
		return nil, fmt.Errorf("gauge mertic %s is not exist", key)
	}
	return &val, nil
}

func (m *Storage) GetCounter(ctx context.Context, key string) (*models.MetricDTO, error) {
	mux.Lock()
	defer mux.Unlock()

	val, ok := m.Counter[key]
	if !ok {
		return nil, fmt.Errorf("counter mertic %s is not exist", key)
	}
	return &val, nil
}

func (m *Storage) GetMetricsList(ctx context.Context) []string {
	mux.Lock()
	defer mux.Unlock()

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

func (m *Storage) GetAllMetrics() []models.MetricDTO {
	mux.Lock()
	defer mux.Unlock()

	metrics := make([]models.MetricDTO, 0)

	for _, metric := range m.Gauge {
		metrics = append(metrics, metric)
	}

	for _, metric := range m.Counter {
		metrics = append(metrics, metric)
	}

	defer m.clearCounter()

	return metrics
}

func (m *Storage) Marshal() ([]byte, error) {
	mux.Lock()
	defer mux.Unlock()

	data, err := json.Marshal(&m)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metrics: %w", err)
	}

	return data, nil
}

func (m *Storage) Restore(data []byte) error {
	mux.Lock()
	defer mux.Unlock()

	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("cannot unmarshal metrics: %w", err)
	}

	return nil
}

func (m *Storage) PingStorage(context.Context) error {
	return nil
}

func (m *Storage) Close() {}
