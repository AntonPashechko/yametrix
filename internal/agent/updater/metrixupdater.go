package updater

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage"
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

type updateMetricsWorker struct {
	storage storage.MetricsStorage
}

func (m *updateMetricsWorker) Work() error {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	/*Делаем json, что бы было убоднее пройтись по нужным метрикам*/
	jMetrics, err := json.Marshal(mem)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %w", err)
	}

	var fields map[string]interface{}
	err = json.Unmarshal(jMetrics, &fields)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json: %w", err)
	}

	for _, gaugeName := range RuntimeGaugesName {
		m.storage.SetGauge(models.NewGaugeMetric(gaugeName, fields[gaugeName].(float64)))
	}

	m.storage.AddCounter(models.NewCounterMetric(pollCount, 1))
	m.storage.SetGauge(models.NewGaugeMetric(randomValue, randFloats()))

	return nil
}

func NewUpdateMetricsWorker(storage storage.MetricsStorage) scheduler.RecurringWorker {
	return &updateMetricsWorker{storage: storage}
}
