package client

import "github.com/AntonPashechko/yametrix/internal/agent/scheduler"

type SendMetrixWorker struct {
	client HTTPClient
}

func (w *SendMetrixWorker) Work() error {
	return w.client.Send()
}

func NewSendMetrixWorker(client HTTPClient) scheduler.RecurringWorker {
	return &SendMetrixWorker{client: client}
}
