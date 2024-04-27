package engine

import (
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"sync"
	"time"
)

type Engine struct {
	scripts     map[string]*scriptHost
	cron        *gocron.Scheduler
	logger      common.Logger
	reader      dynamic.Reader
	kafkaClient *kafkaClient
	m           sync.Mutex
	jsConfig    static.JsConfig
	parallel    bool
}

func New(reader dynamic.Reader, app *runtime.App, jsConfig static.JsConfig, parallel bool) *Engine {
	return &Engine{
		scripts:     make(map[string]*scriptHost),
		cron:        gocron.NewScheduler(time.UTC),
		logger:      log.StandardLogger(),
		reader:      reader,
		kafkaClient: newKafkaClient(app),
		jsConfig:    jsConfig,
		parallel:    parallel,
	}
}

func (e *Engine) AddScript(cfg *dynamic.Config) error {
	if _, ok := cfg.Data.(*script.Script); !ok {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	e.remove(cfg)

	sh := newScriptHost(cfg, e)
	e.scripts[sh.name] = sh

	if e.parallel {
		go func() {
			err := e.compileAndRun(sh)
			if err != nil {
				log.Error(err)
			}
		}()
	} else {
		return e.compileAndRun(sh)
	}

	return nil
}

func (e *Engine) Run(event string, args ...interface{}) []*common.Action {
	var result []*common.Action
	for _, s := range e.scripts {
		result = append(result, s.RunEvent(event, args...)...)
	}

	return result
}

func (e *Engine) Emit(event string, args ...interface{}) []*common.Action {
	return e.Run(event, args...)
}

func (e *Engine) Start() {
	e.cron.StartAsync()
}

func (e *Engine) Close() {
	e.cron.Stop()
}

func (e *Engine) remove(cfg *dynamic.Config) {
	name := cfg.Info.Path()
	if h, ok := e.scripts[name]; ok {
		log.Debugf("updating script %v", cfg.Info.Path())
		h.close()
		delete(e.scripts, name)
	} else {
		log.Debugf("parsing script %v", cfg.Info.Path())
	}
}

func (e *Engine) compileAndRun(sh *scriptHost) error {
	err := sh.Compile()
	if err != nil {
		return err
	}

	err = sh.Run()
	if err != nil {
		return err
	}

	if sh.CanClose() {
		if e.parallel {
			e.m.Lock()
			defer e.m.Unlock()
		}

		sh.close()
		delete(e.scripts, sh.name)
	}

	return nil
}
