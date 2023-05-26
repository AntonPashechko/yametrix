package sign

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

type signWriter struct {
	w http.ResponseWriter
}

func newSignWriter(w http.ResponseWriter) *signWriter {
	return &signWriter{
		w: w,
	}
}

func (m *signWriter) Header() http.Header {
	return m.w.Header()
}

func (m *signWriter) Write(p []byte) (int, error) {
	sign, err := MetricsSigner.CreateSign(p)
	if err != nil {
		return 0, fmt.Errorf("cannot sign request body: %w", err)
	}

	m.w.Header().Set("HashSHA256", hex.EncodeToString(sign))
	return m.w.Write(p)
}

func (m *signWriter) WriteHeader(statusCode int) {
	m.w.WriteHeader(statusCode)
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// проверяем, что клиент отправил серверу заголовок HashSHA256
		if bodyHash := r.Header.Get("HashSHA256"); bodyHash != `` {
			buf, _ := ioutil.ReadAll(r.Body)

			signValue, err := hex.DecodeString(bodyHash)
			if err != nil {
				logger.Error(fmt.Sprintf("bad request sign value: %s", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := MetricsSigner.VerifySign(buf, signValue); err != nil {
				logger.Error(fmt.Sprintf("bad request signature: %s", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rdr1 := io.NopCloser(bytes.NewBuffer(buf))
			r.Body = rdr1
		}

		sw := newSignWriter(w)

		//Вызов целевого handler
		h.ServeHTTP(sw, r)
	})
}
