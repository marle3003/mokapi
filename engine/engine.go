package engine

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
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
	cfgEvent    static.Event
	jobCounter  *metrics.Counter
	sm          *events.StoreManager
}

func New(reader dynamic.Reader, app *runtime.App, config *static.Config, parallel bool) *Engine {
	return &Engine{
		scripts:     make(map[string]*scriptHost),
		scheduler:   NewDefaultScheduler(),
		logger:      newLogger(log.StandardLogger()),
		reader:      reader,
		kafkaClient: NewKafkaClient(app),
		parallel:    parallel,
		loader:      NewDefaultScriptLoader(config),
		cfgEvent:    config.Event,
		jobCounter:  app.Monitor.JobCounter,
		sm:          app.Events,
	}
}

func NewEngine(opts ...Options) *Engine {
	e := &Engine{
		scripts:   make(map[string]*scriptHost),
		scheduler: NewDefaultScheduler(),
		logger:    newLogger(log.StandardLogger()),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) AddScript(evt dynamic.ConfigEvent) error {
	if _, ok := evt.Config.Data.(*script.Script); !ok {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	host := newScriptHost(evt.Config, e)
	e.addOrUpdate(host)
	e.scripts[host.name] = host

	log.Infof("executing script %v", evt.Config.Info.Url)
	if e.parallel {
		go func() {
			err := e.run(host)
			if err != nil {
				log.Errorf("error executing script %v: %v", evt.Config.Info.Url, err)
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
		actions := s.RunEvent(event, args...)
		result = append(result, actions...)
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

func (e *Engine) IsLevelEnabled(level string) bool {
	return e.logger.IsLevelEnabled(level)
}
