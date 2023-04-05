package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AntonPashechko/yametrix/internal/agent/metrix"
)

type Action string

const (
	update Action = "update"
)

type MetrixHTTPClient struct {
	runtimemetrix *metrix.RuntimeMetrix
	client        *http.Client
	endpoint      string
}

func NewMetrixClient(runtimemetrix *metrix.RuntimeMetrix, endpoint string) HTTPClient {
	return &MetrixHTTPClient{
		runtimemetrix: runtimemetrix,
		client:        &http.Client{},
		endpoint:      endpoint,
	}
}

func (mhc *MetrixHTTPClient) createURL(action Action, mtype metrix.MetrixType, name string, value string) string {
	urlParts := []string{mhc.endpoint, string(action), string(mtype), name, value}
	return strings.Join(urlParts, "/")
}

func (mhc *MetrixHTTPClient) post(url string) error {
	// пишем запрос
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	response, err := mhc.client.Do(request)
	if err != nil {
		return err
	}

	response.Body.Close()
	return nil
}

func (mhc *MetrixHTTPClient) Send() error {

	gauges, counters := mhc.runtimemetrix.GetMetrix()

	for key, value := range gauges {
		url := mhc.createURL(update, metrix.Gauge, key, fmt.Sprintf("%f", value))
		err := mhc.post(url)
		if err != nil {
			return err
		}
	}

	for key, value := range counters {
		url := mhc.createURL(update, metrix.Counter, key, fmt.Sprintf("%d", value))
		err := mhc.post(url)
		if err != nil {
			return err
		}
	}

	return nil
}
