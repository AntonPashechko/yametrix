// Package updater предназначени для сбора метрик на клиенте.
package updater

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/models"
)

const (
	totalMemory    = "TotalMemory"     // метрика типа TotalMemory
	freeMemory     = "FreeMemory"      // метрика типа FreeMemory
	cpuUtilization = "CPUutilization1" // метрика типа CPUutilization1
)

// AnotherMetricsProducer собирает метрики, которых нет в пакете runtime.
type AnotherMetricsProducer struct {
	tickerTime time.Duration // интервал сбора метрик
}

// NewAnotherMetricsProducer создает экземпляр AnotherMetricsProducer.
func NewAnotherMetricsProducer(cfg *config.Config) *AnotherMetricsProducer {
	return &AnotherMetricsProducer{
		tickerTime: time.Duration(cfg.PollInterval) * time.Second,
	}
}

// produceMetrics собирает another метрики и отправляет из в канал метрик.
func (m *AnotherMetricsProducer) produceMetrics(metricCh chan<- models.MetricDTO) {
	//В 13 ИНКРЕМЕНТЕ В ТЕСТАХ ОТКУДА-ТО ВЫЛЕЗЛИ МЕТРИКИ ВНЕ ПАКЕТА RUNTIME TotalMemory FreeMemory CPUutilization1
	//ДЛЯ ПЕРВЫХ 2х ПОДКЛЮЧИЛ github.com/pbnjay/memory
	metricCh <- models.NewGaugeMetric(totalMemory, float64(memory.TotalMemory()))
	metricCh <- models.NewGaugeMetric(freeMemory, float64(memory.FreeMemory()))

	//ДЛЯ CPUutilization1 - github.com/shirou/gopsutil
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Printf("cannot get cpu utilization: %s\n", err)
	}

	metricCh <- models.NewGaugeMetric(cpuUtilization, percentage[0])
}

// Work определяет работу, запускает по таймеру процедуру сбора метрик пока не пришел сигнал об отмене.
func (m *AnotherMetricsProducer) Work(ctx context.Context, wg *sync.WaitGroup, metricCh chan<- models.MetricDTO) {
	defer wg.Done()

	ticker := time.NewTicker(m.tickerTime)

	for {
		select {
		// выход по ctx
		case <-ctx.Done():
			return
		// собираем метрики, пишем их в канал
		case <-ticker.C:
			m.produceMetrics(metricCh)
		}
	}
}
