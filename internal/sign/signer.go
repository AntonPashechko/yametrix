package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/AntonPashechko/yametrix/internal/logger"
)

var MetricsSigner *Signer

type Signer struct {
	key []byte
}

func Initialize(key []byte) {
	MetricsSigner = &Signer{
		key: key,
	}
}

func (m *Signer) CreateSign(buf []byte) ([]byte, error) {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, m.key)
	_, err := h.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("signature proccess error: %s", err)
	}

	return h.Sum(nil), nil
}

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
