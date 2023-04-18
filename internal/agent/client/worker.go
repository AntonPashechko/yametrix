package client

import "github.com/AntonPashechko/yametrix/internal/agent/scheduler"

type sendMetrixWorker struct {
	client HTTPClient
}

func (w *sendMetrixWorker) Work() error {
	return w.client.Send()
}

func NewSendMetrixWorker(client HTTPClient) scheduler.RecurringWorker {
	return &sendMetrixWorker{client: client}
}
