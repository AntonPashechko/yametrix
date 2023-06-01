package sender

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/sign"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-resty/resty/v2"
)

const (
	updates = "updates"
)

type metricsConsumer struct {
	storage            *memstorage.Storage
	tickerTime         time.Duration
	endpoint           string
	client             *resty.Client
	retriableIntervals []time.Duration
}

func NewMetricsConsumer(cfg *config.Config) *metricsConsumer {
	return &metricsConsumer{
		storage:            memstorage.NewStorage(),
		tickerTime:         time.Duration(cfg.ReportInterval) * time.Second,
		endpoint:           cfg.ServerEndpoint,
		client:             resty.New(),
		retriableIntervals: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second, time.Nanosecond},
	}
}

func (m *metricsConsumer) retriablePost(req *resty.Request, postURL string) error {
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

func (m *metricsConsumer) postMetrics(buf []byte) error {

	//Создали клиента
	req := m.client.R()

	//Проводим контроль целостности, если надо
	if sign.MetricsSigner != nil {
		sign, err := sign.MetricsSigner.CreateSign(buf)
		if err != nil {
			return fmt.Errorf("cannot sign request body: %w", err)
		}

		req.SetHeader("HashSHA256", hex.EncodeToString(sign))
	}

	//Компресим (после расчета для контроля целостности)
	buf, err := compress.GzipCompress(buf)
	if err != nil {
		return fmt.Errorf("cannot compress data: %w", err)
	}

	req.SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf)

	err = m.retriablePost(req, strings.Join([]string{m.endpoint, updates}, "/"))
	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	return nil
}

func (m *metricsConsumer) Work(wg *sync.WaitGroup, ctx context.Context, metricCh <-chan models.MetricDTO) {

	defer wg.Done()

	ticker := time.NewTicker(m.tickerTime)

	for {
		select {
		// выход по ctx
		case <-ctx.Done():
			return
		//Сохораняем приходящие метрики от поставщиков
		case mertic := <-metricCh:
			m.storage.ApplyMetric(ctx, mertic)
		// отправляем накопленые метрики на сервер
		case <-ticker.C:
			metrics := m.storage.GetAllMetrics()

			//В ЗАДАНИИ СКАЗАНО отправлять пустые батчи не нужно; (12 инкремент)
			if len(metrics) == 0 {
				break
			}

			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(metrics); err != nil {
				fmt.Printf("error encoding metrics %s\n", err)
			}

			err := m.postMetrics(buf.Bytes())
			if err != nil {
				fmt.Printf("cannot send metrics batch: %s\n", err)
			}
		}
	}
}
