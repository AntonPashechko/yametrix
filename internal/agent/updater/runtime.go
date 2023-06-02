package updater

import (
	"context"
	"encoding/json"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

const (
	pollCount   = "PollCount"
	randomValue = "RandomValue"

	floatMin = 1.10
	floatMax = 101.98
)

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

type RuntimeMetricsProducer struct {
	tickerTime time.Duration
}

func NewRuntimeMetricsProducer(cfg *config.Config) *RuntimeMetricsProducer {
	return &RuntimeMetricsProducer{
		tickerTime: time.Duration(cfg.PollInterval) * time.Second,
	}
}

func (m *RuntimeMetricsProducer) produceMetrics(metricCh chan<- models.MetricDTO) {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	/*Делаем json, что бы было удобнее пройтись по нужным метрикам*/
	jMetrics, err := json.Marshal(mem)
	if err != nil {
		//return fmt.Errorf("cannot marshal json: %w", err)
	}

	var fields map[string]interface{}
	err = json.Unmarshal(jMetrics, &fields)
	if err != nil {
		//return fmt.Errorf("cannot unmarshal json: %w", err)
	}

	for _, gaugeName := range RuntimeGaugesName {
		metricCh <- models.NewGaugeMetric(gaugeName, fields[gaugeName].(float64))
	}

	metricCh <- models.NewCounterMetric(pollCount, 1)
	metricCh <- models.NewGaugeMetric(randomValue, randFloats())

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

func (m *RuntimeMetricsProducer) Work(wg *sync.WaitGroup, ctx context.Context, metricCh chan<- models.MetricDTO) {
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
