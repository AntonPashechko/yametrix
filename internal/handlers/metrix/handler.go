package metrix

import (
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

const (
	updateGaugeURL   = "/update/gauge/"
	updateCounterURL = "/update/counter/"
)

func getKeyValue(uri string, prefix string) (bool, []string) {
	data := strings.TrimPrefix(uri, prefix)
	if data == "" {
		return false, nil
	}

	keyValue := strings.Split(data, "/")

	if len(keyValue) != 2 {
		return false, nil
	}

	return true, keyValue
}

type Handler struct {
	Storage storage.MertixStorage
}

func (h *Handler) Register(router *http.ServeMux) {
	router.HandleFunc(updateGaugeURL, h.updateGauge)
	router.HandleFunc(updateCounterURL, h.updateCounter)
}

func (h *Handler) updateGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ok, keyValue := getKeyValue(r.RequestURI, updateGaugeURL)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := utils.StrToFloat64(keyValue[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Storage.Set(keyValue[0], value)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ok, keyValue := getKeyValue(r.RequestURI, updateCounterURL)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := utils.StrToInt64(keyValue[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Storage.Add(keyValue[0], value)

	w.WriteHeader(http.StatusOK)
}
