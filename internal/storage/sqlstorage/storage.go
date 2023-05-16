package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
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
	if err := storage.bootstrap(context.TODO()); err != nil {
		return nil, fmt.Errorf("cannot bootstarp db: %w", err)
	}
	return &Storage{conn: conn}, nil
}

// Bootstrap подготавливает БД к работе, создавая необходимые таблицы и индексы
func (m *Storage) bootstrap(ctx context.Context) error {
	// запускаем транзакцию
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// создаём таблицу для хранения метрик
	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS metrics (
            id varchar(128) PRIMARY KEY,
			type varchar(128),
			delta int,
			value double precision
        )
    `)

	// коммитим транзакцию
	return tx.Commit()
}

// Проверяем, что запись о метрике присутствует в таблице
func (m *Storage) isMerticExist(ctx context.Context, metric models.MetricDTO) bool {
	row := m.conn.QueryRowContext(ctx, "SELECT id FROM metrics WHERE id = $1", metric.ID)
	err := row.Scan(&metric.ID)
	return !(err != nil && err == sql.ErrNoRows)
}

// GetGauge implements storage.MetricsStorage
func (m *Storage) getMetricByID(ctx context.Context, id string) (*models.MetricDTO, error) {
	// делаем запрос
	row := m.conn.QueryRowContext(ctx, "SELECT * FROM metrics WHERE id = $1", id)
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
	if m.isMerticExist(ctx, metric) {
		_, err := m.conn.ExecContext(ctx,
			"UPDATE metrics SET delta = delta + $1 WHERE id = $2", metric.Delta, metric.ID)

		if err != nil {
			return nil, fmt.Errorf("cannot update counter metric %s: %w", metric.ID, err)
		}
	} else {
		_, err := m.conn.ExecContext(ctx,
			"INSERT INTO metrics (id, type, delta)"+
				" VALUES($1,$2,$3)", metric.ID, metric.MType, metric.Delta)

		if err != nil {
			return nil, fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
		}
	}

	return m.getMetricByID(ctx, metric.ID)
}

// SetGauge implements storage.MetricsStorage
func (m *Storage) SetGauge(ctx context.Context, metric models.MetricDTO) error {
	//Если метрики с таким именем не существует - вставляем, иначе обновляем
	if m.isMerticExist(ctx, metric) {
		_, err := m.conn.ExecContext(ctx,
			"UPDATE metrics SET value = $1 WHERE id = $2", metric.Value, metric.ID)

		if err != nil {
			return fmt.Errorf("cannot update gauge metric %s: %w", metric.ID, err)
		}
	} else {
		_, err := m.conn.ExecContext(ctx,
			"INSERT INTO metrics (id, type, value)"+
				" VALUES($1,$2,$3)", metric.ID, metric.MType, metric.Value)

		if err != nil {
			return fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
		}
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
func (m *Storage) GetMetricsList(ctx context.Context) []string {
	panic("unimplemented")
}

func (m *Storage) PingStorage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return m.conn.PingContext(ctx)
}

func (m *Storage) Close() {
	m.conn.Close()
}
