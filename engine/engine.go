package engine

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"sync"
)

type Options func(e *Engine)

type Engine struct {
	scripts     map[string]*scriptHost
	scheduler   Scheduler
	logger      common.Logger
	reader      dynamic.Reader
	kafkaClient common.KafkaClient
	m           sync.Mutex
	loader      ScriptLoader
	parallel    bool
}

func New(reader dynamic.Reader, app *runtime.App, config *static.Config, parallel bool) *Engine {
	return &Engine{
		scripts:     make(map[string]*scriptHost),
		scheduler:   NewDefaultScheduler(),
		logger:      log.StandardLogger(),
		reader:      reader,
		kafkaClient: NewKafkaClient(app),
		parallel:    parallel,
		loader:      NewDefaultScriptLoader(config),
	}
}

func NewEngine(opts ...Options) *Engine {
	e := &Engine{
		scripts:   make(map[string]*scriptHost),
		scheduler: NewDefaultScheduler(),
		logger:    log.StandardLogger(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) AddScript(file *dynamic.Config) error {
	if _, ok := file.Data.(*script.Script); !ok {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	host := newScriptHost(file, e)
	e.addOrUpdate(host)
	e.scripts[host.name] = host

	if e.parallel {
		go func() {
			err := e.run(host)
			if err != nil {
				log.Error(err)
			}
		}()
	} else {
		return e.run(host)
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
	e.scheduler.Start()
}

func (e *Engine) Close() {
	for _, s := range e.scripts {
		s.close()
	}
	e.scheduler.Close()
}

func (e *Engine) addOrUpdate(host *scriptHost) {
	if h, ok := e.scripts[host.name]; ok {
		log.Debugf("updating script %v", host.name)
		h.close()
		delete(e.scripts, host.name)
	} else {
		log.Debugf("parsing script %v", host.name)
	}
	e.scripts[host.name] = host
}

func (e *Engine) run(host *scriptHost) error {
	err := host.Run()
	if err != nil {
		return err
	}

	if host.CanClose() {
		if e.parallel {
			e.m.Lock()
			defer e.m.Unlock()
		}

		host.close()
		delete(e.scripts, host.name)
	}

	return nil
}

func (e *Engine) Scripts() int {
	e.m.Lock()
	defer e.m.Unlock()
	return len(e.scripts)
}
