package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type MetricsHandler struct {
	storage storage.MetricsStorage
}

func NewMetricsHandler(storage storage.MetricsStorage) MetricsHandler {
	return MetricsHandler{
		storage: storage,
	}
}

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

func (m *MetricsHandler) errorRespond(w http.ResponseWriter, code int, err error) {
	logger.Error(err.Error())
	w.WriteHeader(code)
}

func (m *MetricsHandler) getAll(w http.ResponseWriter, r *http.Request) {
	list, err := m.storage.GetMetricsList(r.Context())
	if err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set metrics list: %s", err))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, strings.Join(list, ", "))
}

func (m *MetricsHandler) get(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	var err error

	switch mType {
	case models.GaugeType:
		if metric, err := m.storage.GetGauge(r.Context(), name); err == nil {
			w.Write([]byte(utils.Float64ToStr(*metric.Value)))
			return
		}
	case models.CounterType:
		if metric, err := m.storage.GetCounter(r.Context(), name); err == nil {
			w.Write([]byte(utils.Int64ToStr(*metric.Delta)))
			return
		}
	}

	m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %s", err))
}

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
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set gauge: %s", err))
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
				m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot add counter: %s", err))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	m.errorRespond(w, http.StatusNotImplemented, fmt.Errorf("unknown metric type %s", mType))
}

func (m *MetricsHandler) getJSON(w http.ResponseWriter, r *http.Request) {

	metric, err := models.NewMetricFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metric: %s", err))
		return
	}

	var res *models.MetricDTO

	switch metric.MType {
	case models.GaugeType:
		if res, err = m.storage.GetGauge(r.Context(), metric.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %s", err))
			return
		}
	case models.CounterType:
		if res, err = m.storage.GetCounter(r.Context(), metric.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %s", err))
			return
		}
	default:
		m.errorRespond(w, http.StatusNotFound, fmt.Errorf("unknown metric type: %s", metric.MType))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %s", err))
	}
}

func (m *MetricsHandler) updateJSON(w http.ResponseWriter, r *http.Request) {

	metric, err := models.NewMetricFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metric: %s", err))
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
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot set gauge: %s", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metric); err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %s", err))
		}

	case models.CounterType:
		if metric.Delta == nil {
			m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("counter delta is nil"))
			return
		}
		res, err := m.storage.AddCounter(r.Context(), metric)
		if err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot add counter: %s", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("error encoding response: %s", err))
		}

	default:
		m.errorRespond(w, http.StatusNotImplemented, fmt.Errorf("unknown metric type %s", metric.MType))
	}
}

func (m *MetricsHandler) updateBatchJSON(w http.ResponseWriter, r *http.Request) {
	metrics, err := models.NewMetricsFromJSON(r.Body)
	if err != nil {
		m.errorRespond(w, http.StatusBadRequest, fmt.Errorf("cannot decode metrics batch: %s", err))
		return
	}

	data, _ := json.Marshal(metrics)
	logger.Info(string(data))

	if err := m.storage.AcceptMetricsBatch(r.Context(), metrics); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot accept metrics batch: %s", err))
		return
	}
}

func (m *MetricsHandler) pingDB(w http.ResponseWriter, r *http.Request) {

	if err := m.storage.PingStorage(r.Context()); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot ping store: %s", err))
	}
}
