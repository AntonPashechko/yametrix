package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

const (
	pollCount   = "PollCount"
	randomValue = "RandomValue"

	totalMemory    = "TotalMemory"
	freeMemory     = "FreeMemory"
	cpuUtilization = "CPUutilization1"

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
	storage *memstorage.Storage
}

func (m *updateMetricsWorker) Work() error {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	memory.TotalMemory()

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
		m.storage.SetGauge(context.Background(), models.NewGaugeMetric(gaugeName, fields[gaugeName].(float64)))
	}

	m.storage.AddCounter(context.Background(), models.NewCounterMetric(pollCount, 1))
	m.storage.SetGauge(context.Background(), models.NewGaugeMetric(randomValue, randFloats()))

	//В 13 ИНКРЕМЕНТЕ В ТЕСТАХ ОТКУДА-ТО ВЫЛЕЗЛИ МЕТРИКИ ВНЕ ПАКЕТА RUNTIME TotalMemory FreeMemory CPUutilization1
	//ДЛЯ ПЕРВЫХ 2х ПОДКЛЮЧИЛ github.com/pbnjay/memory
	m.storage.SetGauge(context.Background(), models.NewGaugeMetric(totalMemory, float64(memory.TotalMemory())))
	m.storage.SetGauge(context.Background(), models.NewGaugeMetric(freeMemory, float64(memory.FreeMemory())))
	//ДЛЯ CPUutilization1 - github.com/shirou/gopsutil
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("cannot get cpu utilization: %w", err)
	}
	m.storage.SetGauge(context.Background(), models.NewGaugeMetric(cpuUtilization, percentage[0]))

	return nil
}

func NewUpdateMetricsWorker(storage *memstorage.Storage) scheduler.RecurringWorker {
	return &updateMetricsWorker{storage: storage}
}
