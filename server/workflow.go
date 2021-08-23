package server

import (
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	luam "mokapi/lua"
	"mokapi/lua/kafka"
	lualog "mokapi/lua/log"
	"mokapi/lua/mustache"
	"mokapi/lua/yaml"
	"path/filepath"
)

type eventHandler func(workflow *Workflow, event string, args ...interface{}) bool

type Workflow struct {
	Name          string
	EventHandlers map[string]eventHandler
}

type luaScript struct {
	key       string
	workflows []*Workflow
	state     *lua.LState
	logger    *lualog.Log
}

func (s *Server) Run(event string, args ...interface{}) {
	el := &EventLog{}
	for _, s := range s.scripts {
		l := s.run(event, args...)
		el.Workflows = append(el.Workflows, l...)
	}
}

func (s *Server) AddScript(key string, code string) {
	if script, ok := s.scripts[key]; ok {
		script.close()
	}

	s.scripts[key] = newScript(key, code, kafka.NewKafka(s.writeKafkaMessage))
}

func (s *luaScript) run(event string, args ...interface{}) (logs []*WorkflowLog) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("error in script %v: %v", s.key, r)
		}
	}()

	for _, w := range s.workflows {
		l := &WorkflowLog{Name: w.Name}
		s.logger.Handler = func(s string) {
			l.Log = append(l.Log, s)
		}
		for _, h := range w.EventHandlers {
			b := h(w, event, args...)
			if b {
				logs = append(logs, l)
			}
		}
	}

	s.logger.Handler = nil

	return
}

func newScript(key, code string, kafka *kafka.Kafka) *luaScript {
	s := &luaScript{key: key, workflows: make([]*Workflow, 0), logger: lualog.NewLog()}
	s.state = lua.NewState()
	s.state.SetGlobal("script_path", lua.LString(key))
	s.state.SetGlobal("script_dir", lua.LString(filepath.Dir(key)))
	s.state.SetGlobal("dump", luar.New(s.state, luam.Dump))
	s.state.PreloadModule("log", s.logger.Loader)
	s.state.PreloadModule("yaml", yaml.Loader)
	s.state.PreloadModule("kafka", kafka.Loader)
	s.state.PreloadModule("mustache", mustache.Loader)

	newWorkflow := func(name string) *Workflow {
		w := newWorkflow(name)
		s.workflows = append(s.workflows, w)
		return w
	}

	tbl := s.state.NewTable()
	tbl.RawSetH(lua.LString("new"), luar.New(s.state, newWorkflow))
	s.state.SetGlobal("workflow", tbl)

	go func() {
		err := s.state.DoString(code)
		if err != nil {
			log.Errorf("script error %q: %v", key, err)
		}
	}()

	return s
}

func (s *luaScript) close() {
	if s.state != nil && !s.state.IsClosed() {
		s.state.Close()
	}
}

func (w *Workflow) RegisterEvent(event string, handler eventHandler) {
	w.EventHandlers[event] = handler
}

func newWorkflow(name string) *Workflow {
	return &Workflow{
		Name:          name,
		EventHandlers: make(map[string]eventHandler),
	}
}
