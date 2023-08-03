// Пакет sign для контроля целостности запросов.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

// Через этот глобальный объект работает Middleware.
var MetricsSigner *Signer

// Storage хранит ключ подписи и реализует методы подписания и проверки.
type Signer struct {
	key []byte // ключ подписи
}

// Initialize инициализирует синглтон MetricsSigner.
func Initialize(key []byte) {
	MetricsSigner = &Signer{
		key: key,
	}
}

// CreateSign вычисляет зачение подписи HMAC SHA-256.
func (m *Signer) CreateSign(buf []byte) ([]byte, error) {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, m.key)
	_, err := h.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("signature proccess error: %s", err)
	}

	return h.Sum(nil), nil
}

// VerifySign вычисляет зачение подписи HMAC SHA-256 и сразвнивает с переданым(непосредствено проверка подписи).
func (m *Signer) VerifySign(data []byte, signValue []byte) error {
	newSign, err := m.CreateSign(data)
	if err != nil {
		return fmt.Errorf("cannot create sign value: %s", err)
	}

	logger.Info("signValue: %s", hex.EncodeToString(signValue))
	logger.Info("newSign: %s", hex.EncodeToString(newSign))

	if !hmac.Equal(newSign, signValue) {
		return fmt.Errorf("invalid signature: %s", err)
	}

	return nil
}
