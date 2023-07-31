// Пакет memstorage предназначен для реализации хранилища метрик памяти приложения.
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

// Storage реализует интерфейс storage.MetricsStorage и позволяет хранить метрики в памяти приложения.
type Storage struct {
	//ЗАГЛАВНЫЕ ЧТО БЫ СРАБОТАЛ json.Marshal
	Gauge   map[string]models.MetricDTO
	Counter map[string]models.MetricDTO
}

// clearCounter сбрасывает все counter метрики в 0.
func (m *Storage) clearCounter() {
	m.Counter = make(map[string]models.MetricDTO)
}

// NewStore возвращает новый экземпляр inmemory хранилища.
func NewStorage() *Storage {

	ms := &Storage{}
	ms.Gauge = make(map[string]models.MetricDTO)
	ms.Counter = make(map[string]models.MetricDTO)

	return ms
}

// ApplyMetric принимает метрику и сохраняет/модифицирует ее в зависимостиот типа.
func (m *Storage) ApplyMetric(ctx context.Context, metric models.MetricDTO) {
	if metric.MType == models.GaugeType {
		m.SetGauge(ctx, metric)
	} else if metric.MType == models.CounterType {
		m.AddCounter(ctx, metric)
	}
}

// SetGauge добавляет/модифицирует gauge метрику.
func (m *Storage) SetGauge(ctx context.Context, metric models.MetricDTO) error {
	mux.Lock()
	defer mux.Unlock()

	m.Gauge[metric.ID] = metric
	return nil
}

// AddCounter добавляет/модифицирует counter метрику.
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

// AcceptMetricsBatch принимает к добавлению/модификации массив метрик.
func (m *Storage) AcceptMetricsBatch(ctx context.Context, metrics []models.MetricDTO) error {

	for _, metric := range metrics {
		if metric.MType == models.GaugeType {
			m.SetGauge(ctx, metric)
		} else if metric.MType == models.CounterType {
			m.AddCounter(ctx, metric)
		}
	}

	return nil
}

// GetGauge возвращает gauge метрику по имени.
func (m *Storage) GetGauge(ctx context.Context, key string) (*models.MetricDTO, error) {
	mux.Lock()
	defer mux.Unlock()

	val, ok := m.Gauge[key]
	if !ok {
		return nil, fmt.Errorf("gauge mertic %s is not exist", key)
	}
	return &val, nil
}

// GetCounter возвращает сounter метрику по имени.
func (m *Storage) GetCounter(ctx context.Context, key string) (*models.MetricDTO, error) {
	mux.Lock()
	defer mux.Unlock()

	val, ok := m.Counter[key]
	if !ok {
		return nil, fmt.Errorf("counter mertic %s is not exist", key)
	}
	return &val, nil
}

// GetMetricsList возвращает все существующие метрики в виде массива строк.
func (m *Storage) GetMetricsList(ctx context.Context) ([]string, error) {
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

	return list, nil
}

// GetMetricsList возвращает все существующие метрики.
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

// Marshal возвращает все существующие метрики в виде JSON.
func (m *Storage) Marshal() ([]byte, error) {
	mux.Lock()
	defer mux.Unlock()

	data, err := json.Marshal(&m)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metrics: %w", err)
	}

	return data, nil
}

// Restore восстанавливает метрики из JSON.
func (m *Storage) Restore(data []byte) error {
	mux.Lock()
	defer mux.Unlock()

	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("cannot unmarshal metrics: %w", err)
	}

	return nil
}

// PingStorage - для реализации интерфейса storage.MetricsStorage, не делает ничего.
func (m *Storage) PingStorage(context.Context) error {
	return nil
}

// Close - для реализации интерфейса storage.MetricsStorage, не делает ничего.
func (m *Storage) Close() {}
