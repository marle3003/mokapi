package static

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/tls"
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	Log              *MokApiLog        `json:"log" yaml:"log"`
	ConfigFile       string            `json:"-" yaml:"-"`
	Providers        Providers         `json:"providers" yaml:"providers"`
	Api              Api               `json:"api" yaml:"api"`
	RootCaCert       tls.FileOrContent `json:"rootCaCert" yaml:"rootCaCert"`
	RootCaKey        tls.FileOrContent `json:"rootCaKey" yaml:"rootCaKey"`
	Services         Services          `json:"-" yaml:"-"`
	Js               JsConfig          `json:"js" yaml:"js"`
	Configs          Configs           `json:"configs" yaml:"configs" explode:"config"`
	Help             bool              `json:"-" yaml:"-" aliases:"h"`
	GenerateSkeleton interface{}       `json:"-" yaml:"-" flag:"generate-cli-skeleton"`
	Features         []string          `json:"-" yaml:"-" explode:"feature"`
	Version          bool              `json:"-" yaml:"-" aliases:"v"`
	Event            Event             `json:"event" yaml:"event"`
	Certificates     CertificateStore  `json:"certificates" yaml:"certificates"`
	DataGen          DataGen           `json:"data-gen" yaml:"data-gen" name:"data-gen"`
	Args             []string          `json:"args" yaml:"-" aliases:"args"` // positional arguments
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Log = &MokApiLog{Level: "info", Format: "text"}

	cfg.Api.Port = "8080"
	cfg.Api.Dashboard = true
	cfg.Api.Search.Enabled = false

	cfg.Providers.File.SkipPrefix = []string{"_"}
	cfg.Event.Store = map[string]Store{"default": {Size: 100}}
	cfg.DataGen.OptionalProperties = "0.85"
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
	Search    Search
}

type Search struct {
	Enabled bool
	Types   []string
}

type NgramAnalyzer struct {
	Min int
	Max int
}

type FileProvider struct {
	Filenames   []string `explode:"filename"`
	Directories []string `explode:"directory"`
	SkipPrefix  []string `flag:"skip-prefix"`
	Include     []string
}

type GitProvider struct {
	Urls         []string `explode:"url"`
	PullInterval string   `yaml:"pullInterval" name:"pull-interval"`
	TempDir      string   `yaml:"tempDir" name:"temp-dir"`

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
	GitHub *GitHubAuth `yaml:"gitHub,github"`
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
	GlobalFolders []string     `yaml:"globalFolders" flag:"global-folders"`
	Packages      []NpmPackage `explode:"package"`
}

type NpmPackage struct {
	Name string
	// Specifies an allow list of files to include in mokapi
	Files []string `explode:"file"`
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

type Event struct {
	Store map[string]Store
}

type Store struct {
	Size int64
}

type Configs []string

type CertificateStore struct {
	Static []Certificate
}

type Certificate struct {
	Cert tls.FileOrContent `yaml:"cert" json:"cert"`
	Key  tls.FileOrContent `yaml:"key" json:"key"`
}

type DataGen struct {
	OptionalProperties string `yaml:"optionalProperties" json:"optionalProperties"`
}

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

func (c *Config) Parse() error {
	for _, arg := range c.Args {
		u, err := url.Parse(arg)
		if err != nil {
			return err
		}
		switch u.Scheme {
		case "http", "https":
			c.Providers.Http.Urls = append(c.Providers.Http.Urls, u.String())
		case "git+https", "git+http":
			c.Providers.Git.Urls = append(c.Providers.Git.Urls, strings.TrimPrefix(u.String(), "git+"))
		case "npm":
			c.Providers.Npm.Packages = append(c.Providers.Npm.Packages, NpmPackage{Name: u.String()})
		case "":
			c.Providers.File.Filenames = append(c.Providers.File.Filenames, arg)
		default:
			if u.Opaque != "" {
				c.Providers.File.Filenames = append(c.Providers.File.Filenames, arg)
			} else {
				return fmt.Errorf("positional argument is not supported: %v", arg)
			}
		}
	}
	return nil
}

func (d *DataGen) OptionalPropertiesProbability() float64 {
	defaultValue := 0.85

	if d.OptionalProperties == "" {
		return defaultValue
	}

	switch strings.ToLower(d.OptionalProperties) {
	case "always":
		return 1.0
	case "never":
		return 0.0
	case "often":
		return 0.85
	case "sometimes":
		return 0.5
	case "rarely":
		return 0.15
	}
	f, err := strconv.ParseFloat(d.OptionalProperties, 64)
	if err != nil {
		log.Warnf("Invalid data generator optional properties value '%v': using %v", d.OptionalProperties, defaultValue)
		return defaultValue
	}
	return f
}
