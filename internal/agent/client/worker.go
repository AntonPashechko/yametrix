package client

import "github.com/AntonPashechko/yametrix/internal/agent/shaduller"

type SendMetrixWorker struct {
	client HttpClient
}

func (w *SendMetrixWorker) Work() error {
	return w.client.Send()
}

func NewSendMetrixWorker(client HttpClient) shaduller.RecurringWorker {
	return &SendMetrixWorker{client: client}
}
