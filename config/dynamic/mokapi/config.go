package mokapi

import (
	"mokapi/config/tls"
)

//func init() {
//	dynamic.Register("mokapi", &Config{}, func(config *dynamic.Config, r dynamic.ConfigReader) bool {
//		return true
//	})
//}

type Config struct {
	ConfigPath   string `yaml:"-" json:"-"`
	Workflows    []Workflow
	Certificates []Certificate
}

type Workflow struct {
	Name  string
	Steps []Step
	On    Triggers
	Env   map[string]string
	Vars  map[string]string
}

type Triggers []Trigger

type Trigger struct {
	Http     *HttpTrigger
	Smtp     *SmtpTrigger
	Schedule *ScheduleTrigger
}

type HttpTrigger struct {
	Method string
	Path   string
}

type SmtpTrigger struct {
	Received bool
	Login    bool
	Logout   bool
	Address  string
}

type ScheduleTrigger struct {
	Every string
	// Number of iterations, or less than 1 for unlimited
	Iterations int
}

type Step struct {
	Name  string
	Id    string
	Uses  string
	With  map[string]string
	Run   string
	Shell string
	Env   map[string]string
	If    string
}

type Certificate struct {
	CertFile tls.FileOrContent `yaml:"certFile" json:"certFile"`
	KeyFile  tls.FileOrContent `yaml:"keyFile" json:"keyFile"`
}
