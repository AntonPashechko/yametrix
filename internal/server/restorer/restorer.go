package restorer

type MetrixRestorer interface {
	restore() error
	store() error
	Work() error
}
