package storage

type MertixStorage interface {
	SetGauge(string, float64)
	AddCounter(string, int64)
}
