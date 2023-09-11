package sign

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// signWriter обертка над ответом для добавления подписи в заголовок.
type signWriter struct {
	w http.ResponseWriter
}

// newSignWriter создает экземпляр signWriter.
func newSignWriter(w http.ResponseWriter) *signWriter {
	return &signWriter{
		w: w,
	}
}

// Header возвращает заголовки.
func (m *signWriter) Header() http.Header {
	return m.w.Header()
}

// Write вычисляет значени подписи и добавлет заголовок HashSHA256.
func (m *signWriter) Write(p []byte) (int, error) {
	sign, err := MetricsSigner.CreateSign(p)
	if err != nil {
		return 0, fmt.Errorf("cannot sign request body: %w", err)
	}

	m.w.Header().Set("HashSHA256", hex.EncodeToString(sign))
	return m.w.Write(p)
}

// WriteHeader устанавливает код ответа.
func (m *signWriter) WriteHeader(statusCode int) {
	m.w.WriteHeader(statusCode)
}

// Middleware работа по проверке подписи запроса и устновка подписи в ответ.
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// проверяем, что клиент отправил серверу заголовок HashSHA256
		if bodyHash := r.Header.Get("HashSHA256"); bodyHash != `` {
			buf, _ := io.ReadAll(r.Body)

			logger.Info("input body: %s", string(buf))

			signValue, err := hex.DecodeString(bodyHash)
			if err != nil {
				logger.Error(fmt.Sprintf("bad request sign value: %s", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := MetricsSigner.verifySign(buf, signValue); err != nil {
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
