package static

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Configs    Configs `json:"configs" explode:"config"`
	Help       bool
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Log = &MokApiLog{Level: "info", Format: "text"}
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
	Filenames   []string `explode:"filename"`
	Directories []string `explode:"directory"`
	SkipPrefix  []string `flag:"skip-prefix"`
	Include     []string
}

type GitProvider struct {
	Urls         []string `explode:"url"`
	PullInterval string   `yaml:"pullInterval"`
	TempDir      string   `yaml:"tempDir"`

	Repositories []GitRepo `explode:"repository"`
}

type GitRepo struct {
	Url string
	// Specifies an allow list of files to include in mokapi
	Files []string
	// Specifies an array of filenames pr pattern to include in mokapi
	Include      []string
	Auth         *GitAuth
	PullInterval string `yaml:"pullInterval"`
}

type GitAuth struct {
	GitHub *GitHubAuth
}

type GitHubAuth struct {
	AppId          int64             `yaml:"appId"`
	InstallationId int64             `yaml:"installationId"`
	PrivateKey     tls.FileOrContent `yaml:"privateKey"`
}

type HttpProvider struct {
	Urls          []string `explode:"url"`
	PollInterval  string   `yaml:"pollInterval" flag:"poll-interval"`
	PollTimeout   string   `yaml:"pollTimeout" flag:"poll-timeout"`
	Proxy         string
	TlsSkipVerify bool              `yaml:"tlsSkipVerify" flag:"tls-skip-verify"`
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

type Configs []string

func (c *Configs) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '[' {
		return fmt.Errorf("unexpected token %s; expected '['", token)
	}
	for {
		msg := json.RawMessage{}
		err = dec.Decode(&msg)
		if err != nil {
			return err
		}
		*c = append(*c, string(msg))

		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == ']' {
			return nil
		}

	}
}
