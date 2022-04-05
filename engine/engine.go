package engine

import (
	"fmt"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"net/url"
	"strings"
	"time"
)

type Script interface {
	Run() error
	Close()
}

type EventEmitter interface {
	Emit(event string, args ...interface{})
}

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Summary struct {
	Duration time.Duration
	Tags     map[string]string
}

type Engine struct {
	scripts map[string]*scriptHost
	cron    *gocron.Scheduler
	logger  Logger
	reader  common.Reader
}

func New(reader common.Reader) *Engine {
	return &Engine{
		scripts: make(map[string]*scriptHost),
		cron:    gocron.NewScheduler(time.UTC),
		logger:  log.StandardLogger(),
		reader:  reader,
	}
}

func (e *Engine) AddScript(u *url.URL, code string) error {
	log.Infof("parsing %v", u)
	s, err := newScriptHost(u, code, e)
	if err != nil {
		return err
	}

	if h, ok := e.scripts[s.Name]; ok {
		log.Debugf("updating script %v", s.Name)
		h.close()
	}

	err = s.Run()
	if err != nil {
		return err
	}
	e.scripts[s.Name] = s
	return nil
}

func (e *Engine) Run(event string, args ...interface{}) []*Summary {
	var result []*Summary
	for _, s := range e.scripts {
		result = append(result, s.RunEvent(event, args...)...)
	}

	return result
}

func (e *Engine) Emit(event string, args ...interface{}) {
	e.Run(event, args...)
}

func (e *Engine) Start() {
	e.cron.StartAsync()
}

func (e *Engine) Close() {
	e.cron.Stop()
}

func (s *Summary) String() string {
	var sb strings.Builder
	for _, tag := range s.Tags {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(tag)
	}

	return fmt.Sprintf("tags: %v", sb.String())
}
