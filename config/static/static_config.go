package static

type Config struct {
	Log        *MokApiLog
	ConfigFile string
	Providers  Providers
	Api        Api
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Log = &MokApiLog{Level: "error", Format: "default"}
	cfg.Api.Port = "8080"
	cfg.Api.Dashboard = true
	return cfg
}

type MokApiLog struct {
	Level  string
	Format string
}

type Providers struct {
	File FileProvider
}

type Api struct {
	Port      string
	Dashboard bool
}

type FileProvider struct {
	Filename  string
	Directory string
}
