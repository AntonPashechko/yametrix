package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/models"
)

const (
	pollCount   = "PollCount"   // counter метрика PollCount
	randomValue = "RandomValue" // gauge метрика RandomValue

	floatMin = 1.10   // минимальное число для рандомной генерации
	floatMax = 101.98 // максимальное число для рандомной генерации
)

// RuntimeGaugesName содержит массив необходимых runtime метрик
var RuntimeGaugesName = [...]string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
	"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
	"StackSys", "Sys", "TotalAlloc",
}

func randFloats() float64 {
	return floatMin + rand.Float64()*(floatMax-floatMin)
}

// RuntimeMetricsProducer собирает необходимые runtime метрики.
type RuntimeMetricsProducer struct {
	tickerTime time.Duration
}

// NewRuntimeMetricsProducer создает экземпляр RuntimeMetricsProducer.
func NewRuntimeMetricsProducer(cfg *config.Config) *RuntimeMetricsProducer {
	return &RuntimeMetricsProducer{
		tickerTime: time.Duration(cfg.PollInterval) * time.Second,
	}
}

// produceMetrics собирает runtime метрики и отправляет из в канал метрик.
func (m *RuntimeMetricsProducer) produceMetrics(metricCh chan<- models.MetricDTO) {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	/*Делаем json, что бы было удобнее пройтись по нужным метрикам*/
	jMetrics, err := json.Marshal(mem)
	if err != nil {
		fmt.Printf("cannot marshal json: %s\n", err)
	}

	var fields map[string]interface{}
	err = json.Unmarshal(jMetrics, &fields)
	if err != nil {
		fmt.Printf("cannot unmarshal json: %s\n", err)
	}

	for _, gaugeName := range RuntimeGaugesName {
		metricCh <- models.NewGaugeMetric(gaugeName, fields[gaugeName].(float64))
	}

	metricCh <- models.NewCounterMetric(pollCount, 1)
	metricCh <- models.NewGaugeMetric(randomValue, randFloats())
}

// Work определяет работу, запускает по таймеру процедуру сбора метрик пока не пришел сигнал об отмене.
func (m *RuntimeMetricsProducer) Work(ctx context.Context, wg *sync.WaitGroup, metricCh chan<- models.MetricDTO) {
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
