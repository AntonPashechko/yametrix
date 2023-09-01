package updater

import (
	"context"
	"testing"

	"github.com/AntonPashechko/yametrix/internal/models"
)

func BenchmarkProduceRuntimeMetrics(b *testing.B) {
	m := RuntimeMetricsProducer{}
	metricCh := make(chan models.MetricDTO)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context, metricCh <-chan models.MetricDTO) {
		for {
			select {
			// выход по ctx
			case <-ctx.Done():
				return
			case <-metricCh:
			}
		}
	}(ctx, metricCh)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.produceMetrics(metricCh)
	}
}
