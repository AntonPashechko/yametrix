// Package handlers предназначен для реализации обработчиков пользовательских запросов.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

// MetricsHandler реализует методы обработчиков пользовательских запросов по работе с метриками.
type MetricsHandler struct {
	storage storage.MetricsStorage // храниище метрик
}

// NewMetricsHandler создает экземпляр MetricsHandler.
func NewMetricsHandler(storage storage.MetricsStorage) MetricsHandler {
	return MetricsHandler{
		storage: storage,
	}
}

// Register регистрирует маршруты на роутере.
func (m *MetricsHandler) Register(router *chi.Mux) {
	router.Get("/", m.getAll)

	router.Route("/ping", func(router chi.Router) {
		router.Get("/", m.pingDB)
	})

	router.Route("/update", func(router chi.Router) {
		router.Use(restorer.Middleware)
		router.Post("/", m.updateJSON)
		router.Post("/{type}/{name}/{value}", m.update)
	})

	router.Route("/updates", func(router chi.Router) {
		router.Use(restorer.Middleware)
		router.Post("/", m.updateBatchJSON)
	})

	router.Route("/value", func(router chi.Router) {
		router.Post("/", m.getJSON)
		router.Get("/{type}/{name}", m.get)
	})
}

// errorRespond устанавливает код ошибки и вызывает логирование.
func (m *MetricsHandler) errorRespond(w http.ResponseWriter, code int, err error) {
	logger.Error(err.Error())
	w.WriteHeader(code)
}

// getAll возвращает весь список метрик.
func (m *MetricsHandler) getAll(w http.ResponseWriter, r *http.Request) {
	list, err := m.storage.GetMetricsList(r.Context())
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set metrics list: %w", err))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = io.WriteString(w, strings.Join(list, ", "))
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot write response data: %w", err))
	}
}

// get возвращает метрику по имени и типу.
func (m *MetricsHandler) get(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case models.GaugeType:
		metric, err := m.storage.GetGauge(r.Context(), name)
		if err == nil {
			_, err = w.Write([]byte(utils.Float64ToStr(*metric.Value)))
			if err != nil {
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot write data to responce: %w", err))
			}
			return
		}
		m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %w", err))
	case models.CounterType:
		metric, err := m.storage.GetCounter(r.Context(), name)
		if err == nil {
			_, err = w.Write([]byte(utils.Int64ToStr(*metric.Delta)))
			if err != nil {
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot write data to responce: %w", err))
			}
			return
		}
		m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %w", err))
	}
}

// update обновляет значение метрики по имени и типу.
func (m *MetricsHandler) update(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case models.GaugeType:
		if value, err := utils.StrToFloat64(chi.URLParam(r, "value")); err != nil {
			m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("bad gauge value: %s", chi.URLParam(r, "value")))
			return
		} else {
			err := m.storage.SetGauge(r.Context(), models.NewGaugeMetric(name, value))
			if err != nil {
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set gauge: %w", err))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	case models.CounterType:
		if value, err := utils.StrToInt64(chi.URLParam(r, "value")); err != nil {
			m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("bad counter value: %s", chi.URLParam(r, "value")))
			return
		} else {
			_, err := m.storage.AddCounter(r.Context(), models.NewCounterMetric(name, value))
			if err != nil {
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot add counter: %w", err))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	m.errorRespond(w, http.StatusNotImplemented, fmt.Errorf("unknown metric type %s", mType))
}

// getJSON возвращает json представление метрики по имени и типу.
func (m *MetricsHandler) getJSON(w http.ResponseWriter, r *http.Request) {

	metric, err := models.NewMetricFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metric: %w", err))
		return
	}

	var res *models.MetricDTO

	switch metric.MType {
	case models.GaugeType:
		if res, err = m.storage.GetGauge(r.Context(), metric.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %w", err))
			return
		}
	case models.CounterType:
		if res, err = m.storage.GetCounter(r.Context(), metric.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %w", err))
			return
		}
	default:
		m.errorRespond(w, http.StatusNotFound, fmt.Errorf("unknown metric type: %s", metric.MType))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %w", err))
	}
}

// updateJSON обновляет значение метрики по имени и типу из json представления.
func (m *MetricsHandler) updateJSON(w http.ResponseWriter, r *http.Request) {

	metric, err := models.NewMetricFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metric: %w", err))
		return
	}

	switch metric.MType {
	case models.GaugeType:
		if metric.Value == nil {
			m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("gauge value is nil"))
			return
		}
		err := m.storage.SetGauge(r.Context(), metric)
		if err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set gauge: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metric); err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %w", err))
		}

	case models.CounterType:
		if metric.Delta == nil {
			m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("counter delta is nil"))
			return
		}
		res, err := m.storage.AddCounter(r.Context(), metric)
		if err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot add counter: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %w", err))
		}

	default:
		m.errorRespond(w, http.StatusNotImplemented, fmt.Errorf("unknown metric type %s", metric.MType))
	}
}

// updateBatchJSON производит batch обновление списка метрик.
func (m *MetricsHandler) updateBatchJSON(w http.ResponseWriter, r *http.Request) {
	metrics, err := models.NewMetricsFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metrics batch: %w", err))
		return
	}

	if err := m.storage.AcceptMetricsBatch(r.Context(), metrics); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot accept metrics batch: %w", err))
		return
	}
}

// pingDB проверяет работоспособность хранилища метрик.
func (m *MetricsHandler) pingDB(w http.ResponseWriter, r *http.Request) {

	if err := m.storage.PingStorage(r.Context()); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot ping store: %w", err))
	}
}
