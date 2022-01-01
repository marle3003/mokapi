package static

import (
	"mokapi/config/tls"
)

type Config struct {
	Log        *MokApiLog
	ConfigFile string
	Providers  Providers
	Api        Api
	RootCaCert tls.FileOrContent
	RootCaKey  tls.FileOrContent
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
	Git  GitProvider
	Http HttpProvider
}

type Api struct {
	Port      string
	Path      string
	Dashboard bool
}

type FileProvider struct {
	Filename  string
	Directory string
}

type GitProvider struct {
	Url          string
	PullInterval string
}

type HttpProvider struct {
	Url           string
	PollInterval  string
	Proxy         string
	TlsSkipVerify bool
}
