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
	"sync"
	"time"
)

type Summary struct {
	Duration int64
	Tags     map[string]string
}

type Engine struct {
	scripts     map[string]*scriptHost
	cron        *gocron.Scheduler
	logger      common.Logger
	reader      config.Reader
	kafkaClient *kafkaClient
	m           sync.Mutex
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
	if _, ok := cfg.Data.(*script.Script); !ok {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	e.remove(cfg)

	sh := newScriptHost(cfg, e)
	e.scripts[sh.name] = sh

	err := sh.Compile()
	if err != nil {
		return err
	}

	err = sh.Run()
	if err != nil {
		return err
	}

	if sh.CanClose() {
		sh.close()
		delete(e.scripts, sh.name)
	}

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

func (e *Engine) remove(cfg *config.Config) {
	name := getScriptPath(cfg.Url)
	if h, ok := e.scripts[name]; ok {
		log.Debugf("updating script %v", cfg.Url)
		h.close()
		delete(e.scripts, name)
	} else {
		log.Debugf("parsing script %v", cfg.Url)
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
