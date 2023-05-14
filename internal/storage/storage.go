package storage

import "github.com/AntonPashechko/yametrix/internal/models"

type MetricsStorage interface {
	SetGauge(metric models.MetricsDTO)
	AddCounter(metric models.MetricsDTO) models.MetricsDTO

	GetGauge(key string) (models.MetricsDTO, bool)
	GetCounter(key string) (models.MetricsDTO, bool)

	GetMetricsList() []string
	GetAllMetrics() []models.MetricsDTO

	Marshal() ([]byte, error)
	Restore([]byte) error
}
