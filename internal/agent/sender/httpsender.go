package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-resty/resty/v2"
)

const (
	updates = "updates/"
)

type httpSendWorker struct {
	storage  *memstorage.Storage
	endpoint string
	client   *resty.Client
}

func NewHTTPSendWorker(storage *memstorage.Storage, endpoint string) scheduler.RecurringWorker {
	return &httpSendWorker{
		storage,
		endpoint,
		resty.New(),
	}
}

func (m *httpSendWorker) postMetric(url string, buf []byte) error {

	buf, err := compress.GzipCompress(buf)
	if err != nil {
		return fmt.Errorf("cannot compress data: %w", err)
	}

	_, err = m.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf).
		Post(url)

	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	return nil
}

func (m *httpSendWorker) Work() error {
	metrics := m.storage.GetAllMetrics()

	url := strings.Join([]string{m.endpoint, updates}, "/")

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(metrics); err != nil {
		return fmt.Errorf("error encoding metrics %w", err)
	}

	err := m.postMetric(url, buf.Bytes())
	if err != nil {
		return fmt.Errorf("cannot send metrics batch: %w", err)
	}

	/*for _, metric := range metrics {
		buf := new(bytes.Buffer)

		if err := json.NewEncoder(buf).Encode(metric); err != nil {
			return fmt.Errorf("error encoding metric %w", err)
		}

		err := m.postMetric(url, buf.Bytes())
		if err != nil {
			return fmt.Errorf("cannot send metric: %w", err)
		}
	}*/

	return nil
}
