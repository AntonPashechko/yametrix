package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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

type Handler struct {
	storage storage.MetricsStorage
	db      *sql.DB
}

func NewMetricsHandler(storage storage.MetricsStorage, db *sql.DB) Handler {
	return Handler{
		storage: storage,
		db:      db,
	}
}

func (m *Handler) Register(router *chi.Mux) {
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

func (m *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	list := m.storage.GetMetricsList()

	w.Header().Set("Content-Type", "text/html")
	//w.WriteHeader(http.StatusOK)
	io.WriteString(w, strings.Join(list, ", "))
}

func (m *Handler) get(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if metric, ok := m.storage.GetGauge(name); ok {
			w.Write([]byte(utils.Float64ToStr(*metric.Value)))
			return
		}
	case Counter:
		if metric, ok := m.storage.GetCounter(name); ok {
			w.Write([]byte(utils.Int64ToStr(*metric.Delta)))
			return
		}
	}

	logger.Error("NotFound: metric not found %s/%s", mType, name)
	w.WriteHeader(http.StatusNotFound)
}

func (m *Handler) update(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if value, err := utils.StrToFloat64(chi.URLParam(r, "value")); err != nil {
			logger.Error("BadRequest: bad gauge value %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			m.storage.SetGauge(models.NewGaugeMetric(name, value))
			w.WriteHeader(http.StatusOK)
			return
		}
	case Counter:
		if value, err := utils.StrToInt64(chi.URLParam(r, "value")); err != nil {
			logger.Error("BadRequest: bad counter value %s", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			m.storage.AddCounter(models.NewCounterMetric(name, value))
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	logger.Error("NotFound: unknown metric type %s", mType)
	w.WriteHeader(http.StatusNotImplemented)
}

func (m *Handler) getJSON(w http.ResponseWriter, r *http.Request) {

	var req, res models.MetricsDTO
	var ok bool
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.MType {
	case Gauge:
		if res, ok = m.storage.GetGauge(req.ID); !ok {
			logger.Error("NotFound: unknown gauge metric %s", req.ID)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case Counter:
		if res, ok = m.storage.GetCounter(req.ID); !ok {
			logger.Error("NotFound: unknown counter metric %s", req.ID)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		logger.Error("NotFound: unknown metric type %s", req.MType)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Error("error encoding response: %s", err)
	}
}

func (m *Handler) updateJSON(w http.ResponseWriter, r *http.Request) {

	var req models.MetricsDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.MType {
	case Gauge:
		if req.Value == nil {
			logger.Error("BadRequest: gauge value is nil")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.storage.SetGauge(req)
	case Counter:
		if req.Delta == nil {
			logger.Error("BadRequest: counter delta is nil")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req = m.storage.AddCounter(req)
	default:
		logger.Error("NotFound: unknown metric type %s", req.MType)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(req); err != nil {
		logger.Error("error encoding response: %s", err)
	}
}

func (m *Handler) pingDB(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	if err := m.db.PingContext(ctx); err != nil {
		logger.Error("cannot ping db %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
