package updater

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/AntonPashechko/yametrix/internal/agent/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage"
)

const (
	PollCount   = "PollCount"
	RandomValue = "RandomValue"

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

type updateMetrixWorker struct {
	storage storage.MetrixStorage
}

func (m *updateMetrixWorker) Work() error {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	/*Делаем json, что бы было убоднее пройтись по нужным метрикам*/
	jMetrix, err := json.Marshal(mem)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %w", err)
	}

	var fields map[string]interface{}
	err = json.Unmarshal(jMetrix, &fields)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json: %w", err)
	}

	for _, gaugeName := range RuntimeGaugesName {
		m.storage.SetGauge(gaugeName, fields[gaugeName].(float64))
	}

	m.storage.AddCounter(PollCount, 1)
	m.storage.SetGauge(RandomValue, randFloats())

	return nil
}

func NewUpdateMetrixWorker(storage storage.MetrixStorage) scheduler.RecurringWorker {
	return &updateMetrixWorker{storage: storage}
}
