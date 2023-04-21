package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
	"github.com/go-chi/chi/v5"
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

func (h *Handler) Register(router *chi.Mux) {

	//Подключаем middleware логирования
	router.Use(logger.Middleware)

	router.Get("/", h.getAll)
	router.Get("/value/{type}/{name}", h.get)
	router.Post("/update/{type}/{name}/{value}", h.update)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	list := h.Storage.GetMetrixList()

	io.WriteString(w, strings.Join(list, ", "))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if value, ok := h.Storage.GetGauge(name); ok {
			w.Write([]byte(utils.Float64ToStr(value)))
			return
		}
	case Counter:
		if value, ok := h.Storage.GetCounter(name); ok {
			w.Write([]byte(utils.Int64ToStr(value)))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case Gauge:
		if value, err := utils.StrToFloat64(chi.URLParam(r, "value")); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			h.Storage.SetGauge(name, value)
			w.WriteHeader(http.StatusOK)
			return
		}
	case Counter:
		if value, err := utils.StrToInt64(chi.URLParam(r, "value")); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			h.Storage.AddCounter(name, value)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusNotImplemented)
}
