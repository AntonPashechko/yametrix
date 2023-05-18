package sender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-resty/resty/v2"
)

const (
	updates = "updates/"
)

type httpSendWorker struct {
	storage            *memstorage.Storage
	endpoint           string
	client             *resty.Client
	retriableIntervals []time.Duration
}

func NewHTTPSendWorker(storage *memstorage.Storage, endpoint string) scheduler.RecurringWorker {
	return &httpSendWorker{
		storage,
		endpoint,
		resty.New(),
		[]time.Duration{time.Second, 3 * time.Second, 5 * time.Second, time.Nanosecond},
	}
}

func (m *httpSendWorker) retriablePost(req *resty.Request, postURL string) error {
	var err error
	var urlErr *url.Error

	for _, interval := range m.retriableIntervals {
		_, err = req.Post(postURL)
		if err == nil {
			return nil
		}

		if !errors.As(err, &urlErr) {
			break
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("cannot retriable post metric: %w", err)
}

func (m *httpSendWorker) postMetric(url string, buf []byte) error {

	buf, err := compress.GzipCompress(buf)
	if err != nil {
		return fmt.Errorf("cannot compress data: %w", err)
	}

	req := m.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf)

	err = m.retriablePost(req, url)
	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	return nil
}

func (m *httpSendWorker) Work() error {
	metrics := m.storage.GetAllMetrics()

	//В ЗАДАНИИ СКАЗАНО отправлять пустые батчи не нужно; (12 инкремент)
	if len(metrics) == 0 {
		return fmt.Errorf("metrics is empty")
	}

	url := strings.Join([]string{m.endpoint, updates}, "/")

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(metrics); err != nil {
		return fmt.Errorf("error encoding metrics %w", err)
	}

	err := m.postMetric(url, buf.Bytes())
	if err != nil {
		return fmt.Errorf("cannot send metrics batch: %w", err)
	}

	return nil
}
