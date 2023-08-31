package scheduler

import (
	"time"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// Scheduler - планировщик и исполнитель задачи - по тику в заданном интервале запускаем работу у контролируемого объекта.
type Scheduler struct {
	ticker *time.Ticker    // интервал времени выполнения задачи.
	worker RecurringWorker // выполняемая задача
	done   chan struct{}   // канал для остановки работы
}

// NewScheduler создает экземпляр Scheduler.
func NewScheduler(d int64, worker RecurringWorker) Scheduler {
	return Scheduler{
		ticker: time.NewTicker(time.Duration(d) * time.Second),
		done:   make(chan struct{}),
		worker: worker,
	}
}

// Start запускает работу планировщика - ждем таймер и выполняем работу, или закрываемся
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

// Stop останавливает работу планировщика - останавливает таймер и кидает сообщение в канал остановки
func (m Scheduler) Stop() {
	m.ticker.Stop()
	m.done <- struct{}{}
}
