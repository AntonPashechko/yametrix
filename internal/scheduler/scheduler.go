package scheduler

import (
	"time"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

/*Планировщик и исполнитель задачи - по тику в заданном интервале запускаем работу у контролируемого объекта*/
type Scheduler struct {
	ticker *time.Ticker
	worker RecurringWorker
	done   chan bool
}

func NewScheduler(d int64, worker RecurringWorker) Scheduler {
	return Scheduler{
		ticker: time.NewTicker(time.Duration(d) * time.Second),
		done:   make(chan bool),
		worker: worker,
	}
}

func (m Scheduler) Start() {
	for {
		select {
		case <-m.ticker.C:
			err := m.worker.Work()
			if err != nil {
				logger.Error("ERROR: cannot do the work: %s", err)
			}
		case <-m.done:
			return
		}
	}
}

func (m Scheduler) Stop() {
	m.ticker.Stop()
	m.done <- true
}
