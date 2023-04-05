package metrix

import "github.com/AntonPashechko/yametrix/internal/agent/shaduller"

type UpdateMetrixWorker struct {
	metrix *RuntimeMetrix
}

func (w *UpdateMetrixWorker) Work() error {
	return w.metrix.Update()
}

func NewUpdateMetrixWorker(metrix *RuntimeMetrix) shaduller.RecurringWorker {
	return &UpdateMetrixWorker{metrix: metrix}
}
