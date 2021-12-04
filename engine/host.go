package engine

import (
	"fmt"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/readers/file"
	"mokapi/js"
	"mokapi/lua"
	"path/filepath"
	"time"
)

type eventHandler struct {
	handler func(args ...interface{}) (bool, error)
	tags    map[string]string
}

type scriptHost struct {
	name   string
	engine *Engine
	script Script
	jobs   map[int]*gocron.Job
	events map[string][]*eventHandler
	cwd    string
}

func newScriptHost(key, src string, e *Engine) (*scriptHost, error) {
	sh := &scriptHost{
		name:   key,
		engine: e,
		jobs:   make(map[int]*gocron.Job),
		events: make(map[string][]*eventHandler),
		cwd:    filepath.Dir(key),
	}

	var err error
	switch filepath.Ext(key) {
	case ".js":
		if sh.script, err = js.New(key, src, sh); err != nil {
			return nil, err
		}
	case ".lua":
		if sh.script, err = lua.New(key, src, sh); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported script %v", key)
	}

	return sh, nil
}

func (sh *scriptHost) Run() error {
	return sh.script.Run()
}

func (sh *scriptHost) RunEvent(event string, args ...interface{}) []*Summary {
	var result []*Summary
	for _, eh := range sh.events[event] {
		s := &Summary{
			Tags: eh.tags,
		}
		start := time.Now()

		if b, err := eh.handler(args...); err != nil {
			log.Errorf("unable to execute event handler: %v", err)
		} else if !b {
			continue
		}

		s.Duration = time.Now().Sub(start)
		result = append(result, s)
	}
	return result
}

func (sh *scriptHost) Every(every string, handler func(), times int) (int, error) {
	sh.engine.cron.Every(every)

	if times >= 0 {
		sh.engine.cron.LimitRunsTo(times)
	}

	j, err := sh.engine.cron.Do(handler)
	if err != nil {
		return -1, err
	}

	id := len(sh.jobs)
	sh.jobs[id] = j

	return id, nil
}

func (sh *scriptHost) Cron(expr string, handler func(), times int) (int, error) {
	sh.engine.cron.Cron(expr)

	if times >= 0 {
		sh.engine.cron.LimitRunsTo(times)
	}

	j, err := sh.engine.cron.Do(handler)
	if err != nil {
		return -1, err
	}

	id := len(sh.jobs)
	sh.jobs[id] = j

	return id, nil
}

func (sh *scriptHost) Cancel(jobId int) error {
	if j, ok := sh.jobs[jobId]; !ok {
		return fmt.Errorf("job not defined")
	} else {
		sh.engine.cron.RemoveByReference(j)
		return nil
	}
}

func (sh *scriptHost) On(event string, handler func(args ...interface{}) (bool, error), tags map[string]string) {
	h := &eventHandler{
		handler: handler,
		tags: map[string]string{
			"name":  sh.name,
			"event": event,
		},
	}

	for k, v := range tags {
		h.tags[k] = v
	}

	sh.events[event] = append(sh.events[event], h)
}

func (sh *scriptHost) close() {
	for _, j := range sh.jobs {
		sh.engine.cron.RemoveByReference(j)
	}
	sh.jobs = nil

	sh.script.Close()
}

func (sh *scriptHost) Info(args ...interface{}) {
	sh.engine.logger.Info(args...)
}

func (sh *scriptHost) Warn(args ...interface{}) {
	sh.engine.logger.Warn(args...)
}

func (sh *scriptHost) Error(args ...interface{}) {
	sh.engine.logger.Error(args...)
}

func (sh *scriptHost) OpenFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(sh.cwd, path)
	}

	u, err := file.ParseUrl(path)
	if err != nil {
		return "", err
	}

	f, err := sh.engine.reader.Read(u, common.AsPlaintext())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", f.Data), nil
}
