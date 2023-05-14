package updater

import (
	"testing"

	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateMetricsWorker(t *testing.T) {
	type args struct {
		storage storage.MetricsStorage
	}
	tests := []struct {
		name string
	}{
		{"createupdateMetricsWorker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upd := NewUpdateMetricsWorker(memstorage.NewMemStorage())
			assert.NotEmpty(t, upd)
		})
	}
}

func Test_updateMetricsWorker_Work(t *testing.T) {

	storage := memstorage.NewMemStorage()
	updateWorker := NewUpdateMetricsWorker(storage)

	tests := []struct {
		name    string
		worker  scheduler.RecurringWorker
		wantErr bool
	}{
		{
			name:    "SimpleUpdate",
			worker:  updateWorker,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.worker.Work()

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}
