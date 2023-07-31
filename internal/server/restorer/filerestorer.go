package restorer

import (
	"fmt"
	"os"

	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

type FileRestorer struct {
	storeFileName string              //Имя файла для синхронизации данных
	storage       *memstorage.Storage //Хранилище метрик
}

func NewFileRestorer(storage *memstorage.Storage, path string) FileRestorer {

	return FileRestorer{
		storeFileName: path,
		storage:       storage,
	}
}

func (m *FileRestorer) restore() error {
	data, err := os.ReadFile(m.storeFileName)
	if err != nil {
		return fmt.Errorf("cannot read store file: %w", err)
	}

	return m.storage.Restore(data)
}

// Сохраняем метрики в файл
func (m *FileRestorer) store() error {
	// получаем JSON формат метрик
	data, err := m.storage.Marshal()
	if err != nil {
		return fmt.Errorf("cannot get metrics: %w", err)
	}
	// сохраняем данные в файл
	return os.WriteFile(m.storeFileName, data, 0666)
}

func (m FileRestorer) Work() error {
	return m.store()
}
