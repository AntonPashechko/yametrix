package metrix

import (
	"io"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/handlers"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
	"github.com/go-chi/chi/v5"
)

const (
	updateURL = "/update"
)

type Handler struct {
	Storage storage.MertixStorage
}

func NewMetrixHandler(storage storage.MertixStorage) handlers.Handler {
	return &Handler{Storage: storage}
}

func (h *Handler) Register(router *chi.Mux) {

	router.Get("/", h.getAll)
	router.Get("/value/{type}/{name}", h.get)
	router.Post("/update/{type}/{name}/{value}", h.update)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	list := h.Storage.GetMetrixList()

	io.WriteString(w, strings.Join(list, ", "))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if value, ok := h.Storage.GetGauge(name); ok {
		w.Write([]byte(utils.Float64ToStr(value)))
		return
	}

	if value, ok := h.Storage.GetCounter(name); ok {
		w.Write([]byte(utils.Int64ToStr(value)))
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch mType {
	case "gauge":
		value, err := utils.StrToFloat64(chi.URLParam(r, "value"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.Storage.SetGauge(name, value)
	case "counter":
		value, err := utils.StrToInt64(chi.URLParam(r, "value"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.Storage.AddCounter(name, value)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}
