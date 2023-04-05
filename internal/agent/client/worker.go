package client

import "github.com/AntonPashechko/yametrix/internal/agent/shaduller"

type SendMetrixWorker struct {
	client HTTPClient
}

func (w *SendMetrixWorker) Work() error {
	return w.client.Send()
}

func NewSendMetrixWorker(client HTTPClient) shaduller.RecurringWorker {
	return &SendMetrixWorker{client: client}
}
