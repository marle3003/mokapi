package static

import "mokapi/config/dynamic"

type Config struct {
	Log        *MokApiLog
	ConfigFile string
	Providers  Providers
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Log = &MokApiLog{Level: "error", Format: "default"}
	return cfg
}

type MokApiLog struct {
	Level  string
	Format string
}

type Providers struct {
	File dynamic.FileProvider
}
