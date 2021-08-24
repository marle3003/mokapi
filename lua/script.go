package lua

import (
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/lua/kafka"
	lualog "mokapi/lua/log"
	"mokapi/lua/mustache"
	"mokapi/lua/yaml"
	"path/filepath"
	"time"
)

type Script struct {
	Key       string
	workflows []*workflow
	state     *lua.LState
	logger    *lualog.Log
}

func (s *Script) Run(event string, args ...interface{}) (logs []*WorkflowLog) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("error in script %v: %v", s.Key, r)
		}
	}()

	for _, w := range s.workflows {
		l := &WorkflowLog{Name: w.Name}
		start := time.Now()
		s.logger.Handler = func(s string) {
			l.Log = append(l.Log, s)
		}
		for _, h := range w.EventHandlers {
			b := h(w, event, args...)
			if b {
				logs = append(logs, l)
			}
		}

		end := time.Now()
		l.Duration = end.Sub(start)
	}

	s.logger.Handler = nil

	return
}

func (s *Script) Close() {
	if s.state != nil && !s.state.IsClosed() {
		s.state.Close()
	}
}

func NewScript(key, code string, kafka *kafka.Kafka, scheduler Scheduler) *Script {
	l := &Script{Key: key, workflows: make([]*workflow, 0), logger: lualog.NewLog()}
	l.state = lua.NewState()
	l.state.SetGlobal("script_path", lua.LString(key))
	l.state.SetGlobal("script_dir", lua.LString(filepath.Dir(key)))
	l.state.SetGlobal("dump", luar.New(l.state, Dump))
	l.state.PreloadModule("log", l.logger.Loader)
	l.state.PreloadModule("yaml", yaml.Loader)
	l.state.PreloadModule("kafka", kafka.Loader)
	l.state.PreloadModule("mustache", mustache.Loader)

	newWorkflow := func(name string) *workflow {
		w := newWorkflow(name, scheduler)
		l.workflows = append(l.workflows, w)
		return w
	}

	tbl := l.state.NewTable()
	tbl.RawSetH(lua.LString("new"), luar.New(l.state, newWorkflow))
	l.state.SetGlobal("workflow", tbl)

	go func() {
		err := l.state.DoString(code)
		if err != nil {
			log.Errorf("script error %q: %v", key, err)
		}
	}()

	return l
}
