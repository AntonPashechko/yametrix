package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	Gauge   string = "gauge"
	Counter string = "counter"
)

type Handler struct {
	Storage storage.MetrixStorage
}

func NewMetrixHandler(storage storage.MetrixStorage) Handler {
	return Handler{Storage: storage}
}

func (m *Handler) Register(router *chi.Mux) {

	//Подключаем middleware логирования
	router.Use(logger.Middleware)

	router.Get("/", m.getAll)
	router.Post("/value/", m.get)
	router.Post("/update/", m.update)
}

func (m *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	list := m.Storage.GetMetrixList()

	io.WriteString(w, strings.Join(list, ", "))
}

func (m *Handler) get(w http.ResponseWriter, r *http.Request) {

	var req models.MetricsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.MType {
	case Gauge:
		if value, ok := m.Storage.GetGauge(req.ID); ok {
			req.SetValue(value)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case Counter:
		if detla, ok := m.Storage.GetCounter(req.ID); ok {
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
		logger.Log.Debug("error encoding response", zap.Error(err))
	}
}

func (m *Handler) update(w http.ResponseWriter, r *http.Request) {

	var req models.MetricsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.MType {
	case Gauge:
		if req.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.Storage.SetGauge(req.ID, *req.Value)
	case Counter:
		if req.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		m.Storage.AddCounter(req.ID, *req.Delta)
		*req.Delta, _ = m.Storage.GetCounter(req.ID)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(req); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
	}
}
