package restorer

import (
	"fmt"
	"os"

	"github.com/AntonPashechko/yametrix/internal/storage"
)

type fileRestorer struct {
	storeFileName string                //Имя файла для синхронизации данных
	storage       storage.MetrixStorage //Хранилище метрик
}

func NewFileRestorer(storage storage.MetrixStorage, path string) MetrixRestorer {

	return &fileRestorer{
		storeFileName: path,
		storage:       storage,
	}
}

func (m *fileRestorer) restore() error {
	data, err := os.ReadFile(m.storeFileName)
	if err != nil {
		return fmt.Errorf("cannot read store file: %w", err)
	}

	return m.storage.Restore(data)
}

// Сохраняем метрики в файл
func (m *fileRestorer) store() error {
	// получаем JSON формат метрик
	data, err := m.storage.Marshal()
	if err != nil {
		return fmt.Errorf("cannot get metrics: %w", err)
	}
	// сохраняем данные в файл
	return os.WriteFile(m.storeFileName, data, 0666)
}

func (m *fileRestorer) Work() error {
	return m.store()
}
