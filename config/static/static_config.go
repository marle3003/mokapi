package static

import (
	"mokapi/config/tls"
	"strings"
)

type Config struct {
	Log        *MokApiLog
	ConfigFile string
	Providers  Providers
	Api        Api
	RootCaCert tls.FileOrContent
	RootCaKey  tls.FileOrContent
	Services   Services
	Js         JsConfig
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Log = &MokApiLog{Level: "info", Format: "default"}
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
	Base      string
	Dashboard bool
}

type FileProvider struct {
	Filename   string
	Directory  string
	SkipPrefix []string
}

type GitProvider struct {
	Url          string
	PullInterval string `yaml:"pullInterval"`
	Directories  []string
}

type HttpProvider struct {
	Url           string
	PollInterval  string `yaml:"pollInterval"`
	PollTimeout   string `yaml:"pollTimeout"`
	Proxy         string
	TlsSkipVerify bool `yaml:"tlsSkipVerify"`
}

type Services map[string]*Service

func (s Services) GetByName(name string) *Service {
	key := strings.ReplaceAll(name, " ", "-")
	key = strings.ToLower(key)
	return s[key]
}

type Service struct {
	Config ServiceConfig
	Http   *HttpService
}

type ServiceConfig struct {
	File string
	Url  string
}

type HttpService struct {
	Servers []HttpServer
}

type HttpServer struct {
	Url string
}

type JsConfig struct {
	GlobalFolders []string
}
