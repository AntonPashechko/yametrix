package updater

import (
	"context"
	"sync"
	"time"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

const (
	totalMemory    = "TotalMemory"
	freeMemory     = "FreeMemory"
	cpuUtilization = "CPUutilization1"
)

type AtherMetricsProducer struct {
	tickerTime time.Duration
}

func NewAtherMetricsProducer(cfg *config.Config) *AtherMetricsProducer {
	return &AtherMetricsProducer{
		tickerTime: time.Duration(cfg.PollInterval) * time.Second,
	}
}

func (m *AtherMetricsProducer) produceMetrics(metricCh chan<- models.MetricDTO) {
	//В 13 ИНКРЕМЕНТЕ В ТЕСТАХ ОТКУДА-ТО ВЫЛЕЗЛИ МЕТРИКИ ВНЕ ПАКЕТА RUNTIME TotalMemory FreeMemory CPUutilization1
	//ДЛЯ ПЕРВЫХ 2х ПОДКЛЮЧИЛ github.com/pbnjay/memory
	metricCh <- models.NewGaugeMetric(totalMemory, float64(memory.TotalMemory()))
	metricCh <- models.NewGaugeMetric(freeMemory, float64(memory.FreeMemory()))

	//ДЛЯ CPUutilization1 - github.com/shirou/gopsutil
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		//return fmt.Errorf("cannot get cpu utilization: %w", err)
	}

	metricCh <- models.NewGaugeMetric(cpuUtilization, percentage[0])
}

func (m *AtherMetricsProducer) Work(wg *sync.WaitGroup, ctx context.Context, metricCh chan<- models.MetricDTO) {
	defer wg.Done()

	ticker := time.NewTicker(m.tickerTime)

	for {
		select {
		// если канал doneCh закрылся, выходим из горутины
		case <-ctx.Done():
			return
		// если doneCh не закрыт, отправляем результат вычисления в канал результата
		case <-ticker.C:
			m.produceMetrics(metricCh)
		}
	}
}
