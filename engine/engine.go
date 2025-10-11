package engine

import (
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"sync"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
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
	store       *Store
}

type Store struct {
	data map[string]any
	mu   sync.RWMutex
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
		store:       &Store{data: make(map[string]any)},
	}
}

func NewEngine(opts ...Options) *Engine {
	e := &Engine{
		scripts:   make(map[string]*scriptHost),
		scheduler: NewDefaultScheduler(),
		logger:    newLogger(log.StandardLogger()),
		store:     &Store{data: make(map[string]any)},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) AddScript(evt dynamic.ConfigEvent) error {
	if _, ok := evt.Config.Data.(string); !ok {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	host := newScriptHost(evt.Config, e)
	e.addOrUpdate(host)
	e.scripts[host.name] = host

	if e.parallel {
		go func() {
			err := e.run(host)
			if err != nil {
				if errors.Is(err, UnsupportedError) {
					return
				}
				log.Errorf("error executing script %v: %v", evt.Config.Info.Url, err)
			}
		}()
	} else {
		err := e.run(host)
		if errors.Is(err, UnsupportedError) {
			return nil
		}
		return err
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

func (s *Store) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}
