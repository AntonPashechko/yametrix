package updater

import (
	"testing"

	"github.com/AntonPashechko/yametrix/internal/agent/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateMetrixWorker(t *testing.T) {
	type args struct {
		storage storage.MetrixStorage
	}
	tests := []struct {
		name string
	}{
		{"createUpdateMetrixWorker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upd := NewUpdateMetrixWorker(memstorage.NewMemStorage())
			assert.NotEmpty(t, upd)
		})
	}
}

func Test_updateMetrixWorker_Work(t *testing.T) {

	storage := memstorage.NewMemStorage()
	updateWorker := NewUpdateMetrixWorker(storage)

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
