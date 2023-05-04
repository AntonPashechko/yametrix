package config

type Config struct {
	Endpoint      string
	LogLevel      string
	StoreInterval uint64 //0 - синхронная запись
	StorePath     string
	Restore       bool
}
