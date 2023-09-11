// Package trustedsubnets для проверки доверия к IP клиента
package trustedsubnets

import (
	"fmt"
	"net"
	"net/http"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// MetricsSubnetChecker - глобальный объект через который работает Middleware проверки доверия к IP клиента.
var MetricsSubnetChecker *subnetChecker

// subnetChecker хранит диапазон crid и метод проверки IP клиента на вхождение в этот диапазон.
type subnetChecker struct {
	subnet *net.IPNet
}

func (m *subnetChecker) checkIp(clientip string) bool {
	ip := net.ParseIP(clientip)
	return m.subnet.Contains(ip)
}

// Initialize инициализирует синглтон MetricsSubnetChecker.
func Initialize(cidr string) error {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("cannot parse CIDR: %w", err)
	}

	MetricsSubnetChecker = &subnetChecker{
		subnet: subnet,
	}

	return nil
}

// Middleware работа по проверке доверия к IP клиента.
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Если инициализирован ключ расшифровывания - расшифровываем входящие данные
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
