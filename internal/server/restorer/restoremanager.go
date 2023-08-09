package restorer

import (
	"sync"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	config "github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

type Manager struct {
	restorer  FileRestorer
	scheduler scheduler.Scheduler
}

func (m *Manager) store() {
	//Если синхронная запись и шедулер не запущен
	//По другому тут не проверить, твой вариант с nil не cработает, а я не хочу иметь здесь указатель
	if m.scheduler == (scheduler.Scheduler{}) {
		err := m.restorer.store()
		if err != nil {
			logger.Error("cannot store metrics: %s", err)
		}
	}
}

func (m *Manager) shutdown() {
	//Стопаем если вообще был запущен
	//По другому тут не проверить, твой вариант с nil не работает, а я не хочу иметь здесь указатель
	if m.scheduler != (scheduler.Scheduler{}) {
		m.scheduler.Stop()
	}
}

var instance *Manager
var once sync.Once

func Initialize(storage *memstorage.Storage, cfg *config.Config) {
	//Ресторер используем как синглтон, потому тут я применяю sync.Once, считаю эту конструкцию наиболее подходящей для задачи инициализации синглтона
	once.Do(func() {
		//Если имя файла для Store не задано - просто выходим
		if cfg.StorePath == "" {
			return
		}

		restorer := NewFileRestorer(storage, cfg.StorePath)

		//делаем restore если просят
		if cfg.Restore {
			if err := restorer.restore(); err != nil {
				logger.Error("cannot restore metrics from file %s: %s", cfg.StorePath, err)
			}
		}

		var storeScheduler scheduler.Scheduler
		/*Если периодичность сохранения задана - запускаем шедулер*/
		if cfg.StoreInterval != 0 {
			storeScheduler = scheduler.NewScheduler(int64(cfg.StoreInterval), restorer)
			go storeScheduler.Start()
		}

		instance = &Manager{
			restorer:  restorer,
			scheduler: storeScheduler,
		}
	})
}

func Shutdown() {
	if instance != nil {
		instance.shutdown()
	}
}
