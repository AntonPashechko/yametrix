package storage

type MertixStorage interface {
	Set(string, float64)
	Add(string, int64)
}
