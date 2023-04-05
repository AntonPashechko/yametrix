package metrix

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
)

type MetrixType string

const (
	Gauge   MetrixType = "gauge"
	Counter MetrixType = "counter"

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

type RuntimeMetrix struct {
	sync.Mutex

	gauges   map[string]float64
	counters map[string]int64
}

func NewRuntimeMetrix() *RuntimeMetrix {
	rm := RuntimeMetrix{}
	rm.gauges = make(map[string]float64)
	rm.counters = make(map[string]int64)

	rm.counters[PollCount] = 0
	rm.gauges[RandomValue] = randFloats()

	return &rm
}

func (rm *RuntimeMetrix) GetMetrix() (map[string]float64, map[string]int64) {
	rm.Lock()
	defer rm.Unlock()

	return rm.gauges, rm.counters
}

func (rm *RuntimeMetrix) Update() error {

	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	/*Делаем json, что бы было убоднее пройтись по нужным метрикам*/
	jMetrix, err := json.Marshal(mem)
	if err != nil {
		return err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(jMetrix, &fields)
	if err != nil {
		return err
	}

	rm.Lock()
	defer rm.Unlock()

	for _, gaugeName := range RuntimeGaugesName {
		if reflect.TypeOf(fields[gaugeName]).Kind() != reflect.Float64 {
			return fmt.Errorf("Bad gauge value type, %s not float64", gaugeName)
		}
		rm.gauges[gaugeName] = fields[gaugeName].(float64)
	}

	rm.counters[PollCount] += 1
	rm.gauges[RandomValue] = randFloats()

	return nil
}
