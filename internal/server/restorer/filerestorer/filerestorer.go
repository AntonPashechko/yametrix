package filerestorer

import (
	"fmt"
	"os"

	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
)

type fileRestorer struct {
	storeFileName string                //Имя файла для синхронизации данных
	storage       storage.MetrixStorage //Хранилище метрик
}

func NewFileRestorer(storage storage.MetrixStorage, path string) restorer.MetrixRestorer {

	return &fileRestorer{
		storeFileName: path,
		storage:       storage,
	}
}

func (m *fileRestorer) Restore() error {
	if m.storeFileName == "" {
		return nil
	}

	data, err := os.ReadFile(m.storeFileName)
	if err != nil {
		return fmt.Errorf("cannot read store file: %w", err)
	}

	return m.storage.Restore(data)
}

// Сохраняем метрики в файл
func (m *fileRestorer) Store() error {
	if m.storeFileName == "" {
		return nil
	}

	// получаем JSON формат метрик
	data, err := m.storage.Marhal()
	if err != nil {
		return fmt.Errorf("cannot get metrics: %w", err)
	}
	// сохраняем данные в файл
	return os.WriteFile(m.storeFileName, data, 0666)
}

func (m *fileRestorer) Work() error {
	return m.Store()
}
