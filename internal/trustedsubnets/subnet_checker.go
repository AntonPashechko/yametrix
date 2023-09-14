package trustedsubnets

import (
	"fmt"
	"net"
)

// MetricsSubnetChecker - глобальный объект через который работает Middleware проверки доверия к IP клиента.
var MetricsSubnetChecker *subnetChecker

// subnetChecker хранит диапазон CIDR и метод проверки IP клиента на вхождение в этот диапазон.
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
