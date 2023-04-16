package metrix

import "github.com/AntonPashechko/yametrix/internal/agent/scheduler"

type updateMetrixWorker struct {
	metrix *RuntimeMetrix
}

func (w *updateMetrixWorker) Work() error {
	return w.metrix.Update()
}

func NewUpdateMetrixWorker(metrix *RuntimeMetrix) scheduler.RecurringWorker {
	return &updateMetrixWorker{metrix: metrix}
}
