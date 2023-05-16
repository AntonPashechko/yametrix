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

const (
	Gauge   string = "gauge"
	Counter string = "counter"
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
	case Gauge:
		if metric, err := m.storage.GetGauge(r.Context(), name); err == nil {
			w.Write([]byte(utils.Float64ToStr(*metric.Value)))
			return
		}
	case Counter:
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
	case Gauge:
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
	case Counter:
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

	var req models.MetricDTO
	var res *models.MetricDTO
	var err error

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot decode request JSON body: %s", err))
		return
	}

	switch req.MType {
	case Gauge:
		if res, err = m.storage.GetGauge(r.Context(), req.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %s", err))
			return
		}
	case Counter:
		if res, err = m.storage.GetCounter(r.Context(), req.ID); err != nil {
			m.errorRespond(w, http.StatusNotFound, fmt.Errorf("cannot get metric: %s", err))
			return
		}
	default:
		m.errorRespond(w, http.StatusNotFound, fmt.Errorf("unknown metric type: %s", req.MType))
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
	case Gauge:
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

	case Counter:
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

func (m *MetricsHandler) pingDB(w http.ResponseWriter, r *http.Request) {

	if err := m.storage.PingStorage(r.Context()); err != nil {
		m.errorRespond(w, http.StatusInternalServerError, fmt.Errorf("cannot ping store: %s", err))
	}
}
