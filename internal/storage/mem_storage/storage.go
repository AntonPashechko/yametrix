package memstorage

import "github.com/AntonPashechko/yametrix/internal/storage"

type MemStorage struct {
	Metrix map[string]float64
}

func NewMemStorage() storage.MertixStorage {
	metrix := make(map[string]float64)
	return &MemStorage{Metrix: metrix}
}

func (m *MemStorage) Set(key string, value float64) {
	m.Metrix[key] = value
}

func (m *MemStorage) Add(key string, value int64) {
	if m.Metrix[key] != 0 {
		m.Metrix[key] += float64(value)
	}
}
