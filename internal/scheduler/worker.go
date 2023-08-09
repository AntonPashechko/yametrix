// Package scheduler определяет интерфейс для переодически повторяющейся задачи.
package scheduler

// RecurringWorker - интерфейс для переодически повторяющейся задачи.
type RecurringWorker interface {
	Work() error
}
