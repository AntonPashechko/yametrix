// Package metricsgrpc для реализации grpc сервиса обновления метрик.
package metricsgrpc

import (
	context "context"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var _ MetricsServiceServer = &Service{}

type Service struct {
	storage storage.MetricsStorage
}

func NewService(storage storage.MetricsStorage) Service {
	return Service{
		storage: storage,
	}
}

func (m *Service) UpdateMetrics(ctx context.Context, req *UpdateMetricsReq) (*emptypb.Empty, error) {

	metrics := make([]models.MetricDTO, 0, len(req.Metrics))

	for _, mertic := range req.Metrics {
		switch mertic.Type {
		case MetricType_GAUGE:
			dto := models.MetricDTO{
				ID:    mertic.Id,
				MType: models.GaugeType,
			}
			dto.SetValue(mertic.Value)
			metrics = append(metrics, dto)

		case MetricType_COUNTER:
			dto := models.MetricDTO{
				ID:    mertic.Id,
				MType: models.CounterType,
			}
			dto.SetDelta(mertic.Delta)
			metrics = append(metrics, dto)
		}
	}

	if err := m.storage.AcceptMetricsBatch(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot accept metrics batch: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (m *Service) mustEmbedUnimplementedMetricsServiceServer() {}
