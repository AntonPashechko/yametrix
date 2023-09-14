// Package sender предназначен для отправки метрик на сервер.
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

	"github.com/go-resty/resty/v2"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/encrypt"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/sign"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

const (
	updates = "updates"
)

// metricsConsumer накапливает информацию о метриках и управляет их отправкой на сервер.
type metricsConsumer struct {
	storage            *memstorage.Storage  // хранилище метрик агента
	tickerTime         time.Duration        // таймер для периодической отправки метрик на сервер
	endpoint           string               // эндпоин сервера
	client             *resty.Client        // клиент http
	grpcClient         MetricsServiceClient // grpc клиент
	retriableIntervals []time.Duration      // массив retriable интрервалов для переотправки данных в случае сетевых проблем
	agentIp            string               // IP-адрес хоста агента.
}

// NewMetricsConsumer создает экземпляр metricsConsumer.
func NewMetricsConsumer(cfg *config.Config) (*metricsConsumer, error) {
	if cfg.ServiceType == "http" {

		if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
			cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
		}

		return &metricsConsumer{
			storage:            memstorage.NewStorage(),
			tickerTime:         time.Duration(cfg.ReportInterval) * time.Second,
			endpoint:           cfg.ServerEndpoint,
			client:             resty.New(),
			retriableIntervals: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second, time.Nanosecond},
			agentIp:            cfg.IP,
		}, nil
	} else if cfg.ServiceType == "grpc" {

		conn, err := grpc.Dial(cfg.ServerEndpoint, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("cannot dial grpc: %w", err)
		}

		return &metricsConsumer{
			storage:    memstorage.NewStorage(),
			tickerTime: time.Duration(cfg.ReportInterval) * time.Second,
			grpcClient: NewMetricsServiceClient(conn),
			agentIp:    cfg.IP,
		}, nil
	} else {
		return nil, fmt.Errorf("unknown service type: %s", cfg.ServiceType)
	}
}

// retriablePost реализует повторную отправку данных при наличии ошибок в сети.
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

func (m *metricsConsumer) send(metrics []models.MetricDTO) error {

	if m.client != nil {
		err := m.sendHTTP(metrics)
		if err != nil {
			return fmt.Errorf("cannot send metrics batch by http: %w", err)
		}
	} else {
		err := m.sendGRPC(metrics)
		if err != nil {
			return fmt.Errorf("cannot send metrics batch by grpc: %w", err)
		}
	}

	return nil
}

// postMetrics отправка метрик на сервер.
func (m *metricsConsumer) sendHTTP(metrics []models.MetricDTO) error {

	metricBytes := new(bytes.Buffer)
	if err := json.NewEncoder(metricBytes).Encode(metrics); err != nil {
		return fmt.Errorf("error encoding metrics %w", err)
	}

	buf := metricBytes.Bytes()

	//Создали клиента
	req := m.client.R()

	//Шифруем сообщение, если проинициализирован encryptor
	if encrypt.MetricsEncryptor != nil {
		encryptbuf, err := encrypt.MetricsEncryptor.Encrypt(buf)
		if err != nil {
			return fmt.Errorf("cannot encrypt metrics: %w", err)
		}

		buf = encryptbuf
	}

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
		SetHeader("X-Real-IP", m.agentIp).
		SetBody(buf)

	err = m.retriablePost(req, strings.Join([]string{m.endpoint, updates}, "/"))
	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	return nil
}

// postMetrics отправка метрик на сервер.
func (m *metricsConsumer) sendGRPC(metrics []models.MetricDTO) error {

	gpcrMetrics := make([]*Metric, 0, len(metrics))
	for _, metric := range metrics {
		gpcrMetric := &Metric{
			Id: metric.ID,
		}

		switch metric.MType {
		case models.GaugeType:
			gpcrMetric.Type = MetricType_GAUGE
			gpcrMetric.Value = *metric.Value
		case models.CounterType:
			gpcrMetric.Type = MetricType_GAUGE
			gpcrMetric.Delta = *metric.Delta
		}

		gpcrMetrics = append(gpcrMetrics, gpcrMetric)
	}

	//добавляем IP
	md := metadata.New(map[string]string{"X-Real-IP": m.agentIp})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err := m.grpcClient.UpdateMetrics(ctx, &UpdateMetricsReq{
		Metrics: gpcrMetrics,
	})
	if err != nil {
		return fmt.Errorf("send metrics process error: %w", err)
	}

	return nil
}

// Work управляет процессом получения новых метрик и оправкой их на сервер
func (m *metricsConsumer) Work(ctx context.Context, wg *sync.WaitGroup, metricCh <-chan models.MetricDTO) {

	defer wg.Done()

	ticker := time.NewTicker(m.tickerTime)

	for {
		select {
		// выход по ctx
		case <-ctx.Done():
			return
		//Сохраняем приходящие метрики от поставщиков
		case mertic := <-metricCh:
			m.storage.ApplyMetric(ctx, mertic)
		// отправляем накопленые метрики на сервер
		case <-ticker.C:
			metrics := m.storage.GetAllMetrics()

			//В ЗАДАНИИ СКАЗАНО отправлять пустые батчи не нужно; (12 инкремент)
			if len(metrics) == 0 {
				break
			}

			err := m.send(metrics)
			if err != nil {
				fmt.Printf("cannot send metrics batch: %s\n", err)
			}
		}
	}
}
