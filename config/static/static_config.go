package static

import "mokapi/config/providers/file"

type Config struct {
	Log        *MokApiLog
	ConfigFile string
	Services   map[string]*Service
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

type Service struct {
	ApiProviders *ApiProviders `yaml:"api"`
}

type ApiProviders struct {
	File file.Provider
}
