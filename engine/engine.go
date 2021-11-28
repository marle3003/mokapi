package engine

import (
	"fmt"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type Script interface {
	Run() error
	Close()
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
}

func New() *Engine {
	return &Engine{
		scripts: make(map[string]*scriptHost),
		cron:    gocron.NewScheduler(time.UTC),
		logger:  log.StandardLogger()}
}

func NewWithLogger(logger Logger) *Engine {
	return &Engine{
		scripts: make(map[string]*scriptHost),
		cron:    gocron.NewScheduler(time.UTC),
		logger:  logger}
}

func (e *Engine) AddScript(key, code string) error {
	switch filepath.Ext(key) {
	case ".js":
		if s, err := newScriptHost(key, code, e); err != nil {
			return err
		} else {
			e.scripts[key] = s
			return nil
		}
	case ".lua":

	}

	return fmt.Errorf("unsupported script %v", key)
}

func (e *Engine) Run(event string, args ...interface{}) []*Summary {
	var result []*Summary
	for _, s := range e.scripts {
		result = append(result, s.Run(event, args...)...)
	}

	return result
}
