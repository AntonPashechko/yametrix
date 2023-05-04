package storage

type MetrixStorage interface {
	SetGauge(string, float64)
	AddCounter(string, int64)

	GetGauge(string) (float64, bool)
	GetCounter(string) (int64, bool)

	GetMetrixList() []string
	GetMetrix() (map[string]float64, map[string]int64)

	Marshal() ([]byte, error)
	Restore([]byte) error
}
