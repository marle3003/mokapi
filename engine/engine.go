package engine

import (
	"fmt"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/script"
	"mokapi/engine/common"
	"mokapi/runtime"
	"strings"
	"time"
)

type Summary struct {
	Duration time.Duration
	Tags     map[string]string
}

type Engine struct {
	scripts     map[string]*scriptHost
	cron        *gocron.Scheduler
	logger      common.Logger
	reader      config.Reader
	kafkaClient *kafkaClient
}

func New(reader config.Reader, app *runtime.App) *Engine {
	return &Engine{
		scripts:     make(map[string]*scriptHost),
		cron:        gocron.NewScheduler(time.UTC),
		logger:      log.StandardLogger(),
		reader:      reader,
		kafkaClient: newKafkaClient(app),
	}
}

func (e *Engine) AddScript(cfg *config.Config) error {
	s, ok := cfg.Data.(*script.Script)
	if !ok {
		return nil
	}
	log.Infof("parsing %v", s.Filename)
	sh, err := newScriptHost(cfg, e)
	if err != nil {
		return err
	}

	if old, ok := e.scripts[sh.Name]; ok && cfg.Version <= old.file.Version {
		return nil
	}

	e.remove(sh.Name)

	err = sh.Run()
	if err != nil {
		return err
	}
	e.scripts[sh.Name] = sh

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

func (e *Engine) remove(name string) {
	if h, ok := e.scripts[name]; ok {
		log.Debugf("updating script %v", name)
		h.close()
		delete(e.scripts, name)
	}
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
