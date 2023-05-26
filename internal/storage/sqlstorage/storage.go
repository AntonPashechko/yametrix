package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

const (
	setGaugeSQL       = "INSERT INTO metrics (id, type, value) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET value = $3"
	addCounterSQL     = "INSERT INTO metrics (id, type, delta) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + $3"
	getAllMerticsSQL  = "SELECT * FROM metrics"
	selectMerticsById = "SELECT * FROM metrics WHERE id = $1"
)

var _ storage.MetricsStorage = &Storage{}

// Store реализует интерфейс store.Store и позволяет взаимодействовать с СУБД PostgreSQL
type Storage struct {
	// Поле conn содержит объект соединения с СУБД
	conn *sql.DB
}

// NewStore возвращает новый экземпляр PostgreSQL хранилища
func NewStorage(dns string) (*Storage, error) {
	//Храним метрики в базе postgres
	conn, err := sql.Open("pgx", dns)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection db: %w", err)
	}

	storage := &Storage{conn: conn}
	if err := storage.applyDBMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("cannot bootstarp db: %w", err)
	}
	return &Storage{conn: conn}, nil
}

// Bootstrap подготавливает БД к работе, создавая необходимые таблицы и индексы
func (m *Storage) applyDBMigrations(ctx context.Context) error {
	// запускаем транзакцию
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// создаём таблицу для хранения метрик
	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS metrics (
            id varchar(128) PRIMARY KEY,
			type varchar(128),
			delta bigint,
			value double precision
        )
    `)

	// коммитим транзакцию
	return tx.Commit()
}

// GetGauge implements storage.MetricsStorage
func (m *Storage) getMetricByID(ctx context.Context, id string) (*models.MetricDTO, error) {
	// делаем запрос
	row := m.conn.QueryRowContext(ctx, selectMerticsById, id)
	// готовим переменную для чтения результата

	var metric models.MetricDTO
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value) // разбираем результат
	if err != nil {
		return nil, fmt.Errorf("cannot scan row: %w", err)
	}
	return &metric, nil
}

// AddCounter implements storage.MetricsStorage
func (m *Storage) AddCounter(ctx context.Context, metric models.MetricDTO) (*models.MetricDTO, error) {
	//Если метрики с таким именем не существует - вставляем, иначе обновляем
	_, err := m.conn.ExecContext(ctx, addCounterSQL, metric.ID, metric.MType, metric.Delta)

	if err != nil {
		return nil, fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
	}

	return m.getMetricByID(ctx, metric.ID)
}

// SetGauge implements storage.MetricsStorage
func (m *Storage) SetGauge(ctx context.Context, metric models.MetricDTO) error {
	//Если метрики с таким именем не существует - вставляем, иначе обновляем
	_, err := m.conn.ExecContext(ctx, setGaugeSQL, metric.ID, metric.MType, metric.Value)

	if err != nil {
		return fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
	}

	return nil
}

func (m *Storage) AcceptMetricsBatch(ctx context.Context, metrics []models.MetricDTO) error {

	// начинаем транзакцию
	tx, err := m.conn.Begin()
	if err != nil {
		return fmt.Errorf("cannot start a transaction: %w", err)
	}
	defer tx.Rollback()

	gaugeStmt, err := tx.PrepareContext(ctx, setGaugeSQL)
	if err != nil {
		return fmt.Errorf("cannot create a prepared statement for insert counter metric: %w", err)
	}
	defer gaugeStmt.Close()

	counterStmt, err := tx.PrepareContext(ctx, addCounterSQL)
	if err != nil {
		return fmt.Errorf("cannot create a prepared statement for insert counter metric: %w", err)
	}
	defer counterStmt.Close()

	for _, metric := range metrics {

		if metric.MType == models.GaugeType {

			_, err := gaugeStmt.ExecContext(ctx, metric.ID, metric.MType, metric.Value)
			if err != nil {
				return fmt.Errorf("cannot exec update gauge statement: %w", err)
			}

		} else {
			_, err := counterStmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta)
			if err != nil {
				return fmt.Errorf("cannot exec update gauge statement: %w", err)
			}
		}
	}
	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit the transaction: %w", err)
	}

	return nil
}

func (m *Storage) GetCounter(ctx context.Context, key string) (*models.MetricDTO, error) {
	return m.getMetricByID(ctx, key)
}

// GetGauge implements storage.MetricsStorage
func (m *Storage) GetGauge(ctx context.Context, key string) (*models.MetricDTO, error) {
	return m.getMetricByID(ctx, key)
}

// GetMetricsList implements storage.MetricsStorage
func (m *Storage) GetMetricsList(ctx context.Context) ([]string, error) {

	list := make([]string, 0)

	var metric models.MetricDTO
	rows, err := m.conn.QueryContext(ctx, getAllMerticsSQL)
	if err != nil {
		return nil, fmt.Errorf("cannot query contex: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot scan row: %w", err)
		}

		if metric.MType == models.GaugeType {
			strValue := utils.Float64ToStr(*metric.Value)
			list = append(list, fmt.Sprintf("%s = %s", metric.ID, strValue))
		} else if metric.MType == models.CounterType {
			list = append(list, fmt.Sprintf("%s = %d", metric.ID, *metric.Delta))
		}
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("query rows: %w", err)
	}

	return list, nil
}

func (m *Storage) PingStorage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return m.conn.PingContext(ctx)
}

func (m *Storage) Close() {
	m.conn.Close()
}
