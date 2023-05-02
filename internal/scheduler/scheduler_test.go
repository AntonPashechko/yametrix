package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testWorker struct {
	Complete bool
}

func (t *testWorker) Work() error {
	t.Complete = true
	return nil
}

func TestNewScheduler(t *testing.T) {

	tests := []struct {
		name string
	}{
		{"createScheduler"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := NewScheduler(1, &testWorker{})
			assert.NotEmpty(t, sch)
		})
	}
}

func TestScheduler_Start(t *testing.T) {

	worker := &testWorker{}

	tests := []struct {
		name string
	}{
		{"SchedulerWork"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := NewScheduler(1, worker)
			go sch.Start()

			time.Sleep(2 * time.Second)

			sch.Stop()

			assert.Equal(t, worker.Complete, true)
		})
	}
}
