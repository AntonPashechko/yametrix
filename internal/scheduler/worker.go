package scheduler

/*Интерфейс для переодически повторяющейся задачи*/
type RecurringWorker interface {
	Work() error
}
