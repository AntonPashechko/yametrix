package restorer

type MetrixRestorer interface {
	Restore() error
	Store() error
	Work() error
}
