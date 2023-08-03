// Пакет restorer предназначен для синхронизации inmemory хранилища метрик с файлом на диске.
package restorer

import (
	"fmt"
	"os"

	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

// FileRestorer синхронизирует хранилище метрик с файлом на диске.
type FileRestorer struct {
	storeFileName string              //Имя файла для синхронизации данных
	storage       *memstorage.Storage //Хранилище метрик
}

// NewFileRestorer создает экземпляр FileRestorer.
func NewFileRestorer(storage *memstorage.Storage, path string) FileRestorer {

	return FileRestorer{
		storeFileName: path,
		storage:       storage,
	}
}

// restore востанавливает хранилище из файла.
func (m *FileRestorer) restore() error {
	data, err := os.ReadFile(m.storeFileName)
	if err != nil {
		return fmt.Errorf("cannot read store file: %w", err)
	}

	return m.storage.Restore(data)
}

// store сохраняет метрики в файл.
func (m *FileRestorer) store() error {
	// получаем JSON формат метрик
	data, err := m.storage.Marshal()
	if err != nil {
		return fmt.Errorf("cannot get metrics: %w", err)
	}
	// сохраняем данные в файл
	return os.WriteFile(m.storeFileName, data, 0666)
}

// Work для реализации инфтрефейса RecurringWorker.
func (m FileRestorer) Work() error {
	return m.store()
}
