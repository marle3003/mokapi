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
	Npm  NpmProvider
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
	Urls         []string
	PullInterval string `yaml:"pullInterval"`

	Repositories []GitRepo
}

type GitRepo struct {
	Url string
	// Specifies an allow list of files to include in mokapi
	Files []string
	// Specifies an array of filenames pr pattern to include in mokapi
	Include []string
	Auth    *GitAuth
}

type GitAuth struct {
	GitHub *GitHubAuth
}

type GitHubAuth struct {
	AppId          int64
	InstallationId int64
	PrivateKey     tls.FileOrContent
}

type HttpProvider struct {
	Url           string
	Urls          []string
	PollInterval  string `yaml:"pollInterval"`
	PollTimeout   string `yaml:"pollTimeout"`
	Proxy         string
	TlsSkipVerify bool              `yaml:"tlsSkipVerify"`
	Ca            tls.FileOrContent `yaml:"ca"`
}

type NpmProvider struct {
	GlobalFolders []string `yaml:"globalFolders"`
	Packages      []NpmPackage
}

type NpmPackage struct {
	Name string
	// Specifies an allow list of files to include in mokapi
	Files []string
	// Specifies an array of filenames pr pattern to include in mokapi
	Include []string
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
