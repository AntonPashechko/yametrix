package restorer

import (
	"sync"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	config "github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"go.uber.org/zap"
)

type RestorerType int

const (
	FileRestorer RestorerType = iota + 1
)

type Manager struct {
	restorer  MetrixRestorer
	scheduler scheduler.Scheduler
}

func (m *Manager) store() {
	//Если синхронная запись и шедулер не запущен
	if m.scheduler == (scheduler.Scheduler{}) {
		m.restorer.store()
	}
}

func (m *Manager) shutdown() {
	//Стопаем если вообще был запущен
	if m.scheduler != (scheduler.Scheduler{}) {
		m.scheduler.Stop()
	}
}

var instance *Manager
var once sync.Once

func Initialize(storage storage.MetrixStorage, mType RestorerType, cfg *config.Config) {
	once.Do(func() {
		var restorer MetrixRestorer

		switch mType {
		case FileRestorer:
			restorer = NewFileRestorer(storage, cfg.StorePath)
		default:
			logger.Log.Error("bad restore type")
			return
		}

		//делаем restore если просят
		if cfg.Restore {
			if err := restorer.restore(); err != nil {
				logger.Log.Error("cannot restore metrics", zap.String("file", cfg.StorePath), zap.Error(err))
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
