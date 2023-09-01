package encrypt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// Middleware работа по проверке подписи запроса и устновка подписи в ответ.
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Если инициализирован ключ расшифровывания - расшифровываем входящие данные
		if MetricsDecryptor != nil {
			buf, _ := io.ReadAll(r.Body)

			logger.Info("encrypt body: %s", string(buf))

			message, err := MetricsDecryptor.Decrypt(buf)
			if err != nil {
				logger.Error(fmt.Sprintf("cannot decrypt request body: %s", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(message))
		}

		//Вызов целевого handler
		h.ServeHTTP(w, r)
	})
}
