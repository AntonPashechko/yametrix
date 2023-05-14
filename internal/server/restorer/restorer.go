package restorer

type MetricsRestorer interface {
	restore() error
	store() error
	Work() error
}
