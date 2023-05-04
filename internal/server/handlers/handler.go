package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/AntonPashechko/yametrix/internal/compress"
)

const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

type Handler struct {
	storage storage.MetrixStorage
}

func NewMetrixHandler(storage storage.MetrixStorage) Handler {
	return Handler{storage: storage}
}

func (m *Handler) Register(router *chi.Mux) {

	//Подключаем middleware логирования
	router.Use(logger.Middleware)
	//Подключаем middleware декомпрессии
	router.Use(compress.Middleware)

	router.Get("/", m.getAll)

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
	list := m.storage.GetMetrixList()

	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, strings.Join(list, ", "))
}

func (m *Handler) get(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if value, ok := m.storage.GetGauge(name); ok {
			w.Write([]byte(utils.Float64ToStr(value)))
			return
		}
	case Counter:
		if value, ok := m.storage.GetCounter(name); ok {
			w.Write([]byte(utils.Int64ToStr(value)))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (m *Handler) update(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if value, err := utils.StrToFloat64(chi.URLParam(r, "value")); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			m.storage.SetGauge(name, value)
			w.WriteHeader(http.StatusOK)
			return
		}
	case Counter:
		if value, err := utils.StrToInt64(chi.URLParam(r, "value")); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			m.storage.AddCounter(name, value)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotImplemented)
}

func (m *Handler) getJSON(w http.ResponseWriter, r *http.Request) {

	var req models.MetricsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.MType {
	case Gauge:
		if value, ok := m.storage.GetGauge(req.ID); ok {
			req.SetValue(value)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case Counter:
		if detla, ok := m.storage.GetCounter(req.ID); ok {
			req.SetDelta(detla)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(req); err != nil {
		logger.Log.Error("error encoding response", zap.Error(err))
	}
}

func (m *Handler) updateJSON(w http.ResponseWriter, r *http.Request) {

	var req models.MetricsDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Log.Info(req.ID)

	switch req.MType {
	case Gauge:
		if req.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.storage.SetGauge(req.ID, *req.Value)
	case Counter:
		if req.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		m.storage.AddCounter(req.ID, *req.Delta)
		*req.Delta, _ = m.storage.GetCounter(req.ID)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(req); err != nil {
		logger.Log.Error("error encoding response", zap.Error(err))
	}
}
