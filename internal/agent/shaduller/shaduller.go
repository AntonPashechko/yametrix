package shaduller

import (
	"fmt"
	"time"
)

/*Планировщик и исполнитель задачи - по тику в заданном интервале запускаем работу у контролируемого объекта*/
type Shaduller struct {
	ticker *time.Ticker
	worker RecurringWorker
	ch     chan bool
}

func NewShaduller(d time.Duration, worker RecurringWorker) Shaduller {
	return Shaduller{
		ticker: time.NewTicker(d),
		ch:     make(chan bool),
		worker: worker,
	}
}

func (s Shaduller) Start() {
	defer func() { s.ch <- true }()
	for {
		select {
		case <-s.ticker.C:
			err := s.worker.Work()
			if err != nil {
				fmt.Println(err)
			}
		case <-s.ch:
			return
		}
	}
}

func (s Shaduller) Stop() {
	s.ticker.Stop()
	s.ch <- true
	<-s.ch
}
