package storage

type MertixStorage interface {
	SetGauge(string, float64)
	AddCounter(string, int64)

	GetGauge(string) (float64, bool)
	GetCounter(string) (int64, bool)

	GetMetrixList() []string
}
