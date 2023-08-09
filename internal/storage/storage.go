// Package storage определяет интерфейс MetricsStorage.
package storage

import (
	"context"

	"github.com/AntonPashechko/yametrix/internal/models"
)

// MetricsStorage задает интерфейс для хранилища метрик.
type MetricsStorage interface {
	// SetGauge добавляет/модифицирует метрику типа gauge.
	SetGauge(context.Context, models.MetricDTO) error
	// AddCounter добавляет/модифицирует метрику типа сounter.
	AddCounter(context.Context, models.MetricDTO) (*models.MetricDTO, error)
	// AcceptMetricsBatch принимает к добавлению/модификации массив метрик.
	AcceptMetricsBatch(context.Context, []models.MetricDTO) error
	// GetGauge возвращает gauge метрику по имени.
	GetGauge(context.Context, string) (*models.MetricDTO, error)
	// GetCounter возвращает сounter метрику по имени.
	GetCounter(context.Context, string) (*models.MetricDTO, error)
	// GetMetricsList возвращает все существующие метрики в виде массива строк.
	GetMetricsList(context.Context) ([]string, error)
	// PingStorage проверяет жизнеспособность хранилища метрик.
	PingStorage(context.Context) error
	// Close закрывает хранилище метрик.
	Close()
}
