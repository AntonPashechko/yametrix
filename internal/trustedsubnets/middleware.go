// Package trustedsubnets для проверки доверия к IP клиента
package trustedsubnets

import (
	"net/http"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// Middleware работа по проверке доверия к IP клиента.
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Если инициализирован объект для проверки подсети
		if MetricsSubnetChecker != nil {
			// проверяем, что клиент передал X-Real-IP заголовок, error если его нет
			clientip := r.Header.Get("X-Real-IP")
			if clientip == `` {
				logger.Error("cannot get X-Real-IP header value")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			//Проверяем, что в диапазоне cidr
			if !MetricsSubnetChecker.checkIp(clientip) {
				logger.Error("client ip is not in CIDR range")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		//Вызов целевого handler
		h.ServeHTTP(w, r)
	})
}
