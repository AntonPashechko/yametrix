// Пакет sqlstorage предназначен для реализации хранилища метрик в СУБД PostgreSQL.
package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

// SQL запросы для взаимодействия с хранилищем.
const (
	// Добавление/модификация gauge метрики.
	setGaugeSQL = "INSERT INTO metrics (id, type, value) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET value = $3"
	// Добавление/модификация counter метрики.
	addCounterSQL = "INSERT INTO metrics (id, type, delta) VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + $3"
	// Получение всех метрик.
	getAllMerticsSQL = "SELECT * FROM metrics"
	// Получение метрики по ее имени.
	selectMerticsByID = "SELECT * FROM metrics WHERE id = $1"
	// Batch добавление/модификация gauge метрик.
	setGaugesBatch = "INSERT INTO metrics (id, type, value) VALUES%s ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value"
	// Batch добавление/модификация counter метрик.
	setCountersBatch = "INSERT INTO metrics (id, type, delta) VALUES%s ON CONFLICT (id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta"
)

var _ storage.MetricsStorage = &Storage{}

// Storage реализует интерфейс storage.MetricsStorage и позволяет взаимодействовать с СУБД PostgreSQL.
type Storage struct {
	// Поле conn содержит объект соединения с СУБД
	conn *sql.DB
}

// NewStore возвращает новый экземпляр PostgreSQL хранилища.
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

// ApplyDBMigrations подготавливает БД к работе, создавая необходимые таблицы.
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

// getMetricByID возвращает метрику по ее имени.
func (m *Storage) getMetricByID(ctx context.Context, id string) (*models.MetricDTO, error) {
	// делаем запрос
	row := m.conn.QueryRowContext(ctx, selectMerticsByID, id)
	// готовим переменную для чтения результата

	var metric models.MetricDTO
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value) // разбираем результат
	if err != nil {
		return nil, fmt.Errorf("cannot scan row: %w", err)
	}
	return &metric, nil
}

// AddCounter добавляет/модифицирует counter метрику.
func (m *Storage) AddCounter(ctx context.Context, metric models.MetricDTO) (*models.MetricDTO, error) {
	//Если метрики с таким именем не существует - вставляем, иначе обновляем
	_, err := m.conn.ExecContext(ctx, addCounterSQL, metric.ID, metric.MType, metric.Delta)

	if err != nil {
		return nil, fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
	}

	return m.getMetricByID(ctx, metric.ID)
}

// SetGauge добавляет/модифицирует gauge метрику.
func (m *Storage) SetGauge(ctx context.Context, metric models.MetricDTO) error {
	//Если метрики с таким именем не существует - вставляем, иначе обновляем
	_, err := m.conn.ExecContext(ctx, setGaugeSQL, metric.ID, metric.MType, metric.Value)

	if err != nil {
		return fmt.Errorf("cannot insert gauge metric %s: %w", metric.ID, err)
	}

	return nil
}

// AcceptMetricsBatch принимает к добавлению/модификации массив метрик.
func (m *Storage) AcceptMetricsBatch(ctx context.Context, metrics []models.MetricDTO) error {

	/*Нужно сразу правильно подготовить данные, не должно быть повторяющихся метрик в batch запросе, ON CONFLICT не поможет
	https://pganalyze.com/docs/log-insights/app-errors/U126
	ON CONFLICT поможет только если в базе уже есть такая метрика*/

	gaugesMap := make(map[string]models.MetricDTO, 0)
	countersMap := make(map[string]models.MetricDTO, 0)

	for _, metric := range metrics {

		if metric.MType == models.GaugeType {
			//Тут фиксируем последнюю метрику
			gaugesMap[metric.ID] = metric
		} else {
			//А тут суммируем
			if value, ok := countersMap[metric.ID]; ok {
				value.SetDelta(*value.Delta + *metric.Delta)
			} else {
				countersMap[metric.ID] = metric
			}
		}
	}

	// начинаем транзакцию
	tx, err := m.conn.Begin()
	if err != nil {
		return fmt.Errorf("cannot start a transaction: %w", err)
	}
	defer tx.Rollback()

	if len(gaugesMap) > 0 {
		//Тут составляем запрос для gauges
		gauges := make([]string, 0, len(gaugesMap))
		names := make([]interface{}, 0, len(gaugesMap))
		i := 1
		for _, metric := range gaugesMap {
			gauges = append(gauges, fmt.Sprintf("($%d, 'gauge', %s)", i, utils.Float64ToStr(*metric.Value))) //Без конвертации float скукожится и тесты не проходят
			i++
			names = append(names, metric.ID)
		}

		gaugesReq := fmt.Sprintf(setGaugesBatch, strings.Join(gauges, ","))

		_, err = tx.ExecContext(ctx, gaugesReq, names...)
		if err != nil {
			return fmt.Errorf("cannot exec gauges batch: %w", err)
		}
	}

	if len(countersMap) > 0 {
		//Тут составляем запрос для gauges
		counters := make([]string, 0, len(countersMap))
		names := make([]interface{}, 0, len(countersMap))
		i := 1
		for _, metric := range countersMap {
			counters = append(counters, fmt.Sprintf("($%d, 'counter', %d)", i, *metric.Delta))
			i++
			names = append(names, metric.ID)
		}

		countersReq := fmt.Sprintf(setCountersBatch, strings.Join(counters, ","))

		_, err = tx.ExecContext(ctx, countersReq, names...)
		if err != nil {
			return fmt.Errorf("cannot exec counters batch: %w", err)
		}
	}

	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit the transaction: %w", err)
	}

	return nil
}

// GetCounter возвращает сounter метрику по имени.
func (m *Storage) GetCounter(ctx context.Context, key string) (*models.MetricDTO, error) {
	return m.getMetricByID(ctx, key)
}

// GetGauge возвращает gauge метрику по имени.
func (m *Storage) GetGauge(ctx context.Context, key string) (*models.MetricDTO, error) {
	return m.getMetricByID(ctx, key)
}

// GetMetricsList возвращает все существующие метрики в виде массива строк.
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

// PingStorage проверяет жизнеспособность хранилища.
func (m *Storage) PingStorage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return m.conn.PingContext(ctx)
}

// Close закрывает хранилище метрик.
func (m *Storage) Close() {
	m.conn.Close()
}
