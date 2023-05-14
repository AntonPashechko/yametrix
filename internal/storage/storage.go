package storage

import (
	"context"

	"github.com/AntonPashechko/yametrix/internal/models"
)

type MetricsStorage interface {
	SetGauge(context.Context, models.MetricDTO) error
	AddCounter(context.Context, models.MetricDTO) (*models.MetricDTO, error)

	GetGauge(context.Context, string) (*models.MetricDTO, error)
	GetCounter(context.Context, string) (*models.MetricDTO, error)
	GetMetricsList(context.Context) []string

	PingStorage(context.Context) error
	Close()
}
