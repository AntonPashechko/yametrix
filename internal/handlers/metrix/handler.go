package metrix

import (
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

const (
	updateURL = "/update/"
)

type Handler struct {
	Storage storage.MertixStorage
}

func (h *Handler) Register(router *http.ServeMux) {
	router.HandleFunc(updateURL, h.update)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data := strings.TrimPrefix(r.RequestURI, updateURL)
	if data == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	datas := strings.Split(data, "/")
	if len(datas) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch datas[0] {
	case "gauge":
		value, err := utils.StrToFloat64(datas[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.Storage.Set(datas[0], value)
	case "counter":
		value, err := utils.StrToInt64(datas[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.Storage.Add(datas[0], value)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}
