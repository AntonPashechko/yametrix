package scheduler

import (
	"fmt"
	"time"
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

func (s Scheduler) Start() {
	defer func() { s.done <- true }()
	for {
		select {
		case <-s.ticker.C:
			err := s.worker.Work()
			if err != nil {
				fmt.Println(err)
			}
		case <-s.done:
			return
		}
	}
}

func (s Scheduler) Stop() {
	s.ticker.Stop()
	s.done <- true
	<-s.done
}
