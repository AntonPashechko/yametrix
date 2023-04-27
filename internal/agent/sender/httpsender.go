package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/agent/scheduler"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"

	update = "update/"
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

func (m *httpSendWorker) createURL(mtype string, name string, value string) string {
	urlParts := []string{m.endpoint, update, string(mtype), name, value}
	return strings.Join(urlParts, "/")
}

func (m *httpSendWorker) post(url string, body io.Reader) error {
	// пишем запрос
	request, err := http.NewRequest(http.MethodPost, url, body)
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

	url := strings.Join([]string{m.endpoint, update}, "/")

	buf := new(bytes.Buffer)

	for key, value := range gauges {
		metricsDTO := models.NewMetricsDTO(key, GaugeType, nil, &value)
		if err := json.NewEncoder(buf).Encode(metricsDTO); err != nil {
			return fmt.Errorf("error encoding metric %w", err)
		}

		err := m.post(url, buf)
		if err != nil {
			return fmt.Errorf("cannot send gauge metric: %w", err)
		}
	}

	for key, value := range counters {
		metricsDTO := models.NewMetricsDTO(key, CounterType, &value, nil)
		if err := json.NewEncoder(buf).Encode(metricsDTO); err != nil {
			return fmt.Errorf("error encoding metric %w", err)
		}

		err := m.post(url, buf)
		if err != nil {
			return fmt.Errorf("cannot send counter metric: %w", err)
		}
	}

	return nil
}
