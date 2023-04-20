package sender

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/agent/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage"
)

type MetrixType string

const (
	GaugeType   MetrixType = "gauge"
	CounterType MetrixType = "counter"

	update = "update"
)

type httpSendWorker struct {
	storage  storage.MetrixStorage
	endpoint string
	client   *http.Client
}

func NewHTTPSendWorker(storage storage.MetrixStorage, endpoint string) scheduler.RecurringWorker {
	return &httpSendWorker{
		storage,
		endpoint,
		&http.Client{},
	}
}

func (m *httpSendWorker) createURL(mtype MetrixType, name string, value string) string {
	urlParts := []string{m.endpoint, update, string(mtype), name, value}
	return strings.Join(urlParts, "/")
}

func (m *httpSendWorker) post(url string) error {
	// пишем запрос
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("cannot make http request: %w", err)
	}

	response, err := m.client.Do(request)
	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	response.Body.Close()
	return nil
}

func (m *httpSendWorker) Work() error {
	gauges, counters := m.storage.GetMetrix()

	for key, value := range gauges {
		url := m.createURL(GaugeType, key, fmt.Sprintf("%f", value))
		err := m.post(url)
		if err != nil {
			return fmt.Errorf("cannot send gauge metric: %w", err)
		}
	}

	for key, value := range counters {
		url := m.createURL(CounterType, key, fmt.Sprintf("%d", value))
		err := m.post(url)
		if err != nil {
			return fmt.Errorf("cannot send counter metric: %w", err)
		}
	}

	return nil
}
