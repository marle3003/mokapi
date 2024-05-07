package engine

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/engine/common"
	"mokapi/json/generator"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"
)

type eventHandler struct {
	handler func(args ...interface{}) (bool, error)
	tags    map[string]string
}

type scriptHost struct {
	name string

	engine *Engine

	script common.Script
	jobs   map[int]Job
	events map[string][]*eventHandler
	file   *dynamic.Config

	fakerNodes []*fakerTree
	m          sync.Mutex
}

func newScriptHost(file *dynamic.Config, engine *Engine) *scriptHost {
	sh := &scriptHost{
		name:   file.Info.Path(),
		jobs:   make(map[int]Job),
		events: make(map[string][]*eventHandler),
		file:   file,
		engine: engine,
	}

	return sh
}

func (sh *scriptHost) Name() string {
	return sh.name
}

func (sh *scriptHost) Run() (err error) {
	if sh.script == nil {
		sh.script, err = sh.engine.loader.Load(sh.file, sh)
		if err != nil {
			return
		}
	}
	return sh.script.Run()
}

func (sh *scriptHost) RunEvent(event string, args ...interface{}) []*common.Action {
	var result []*common.Action
	for _, eh := range sh.events[event] {
		s := &common.Action{
			Tags: eh.tags,
		}
		start := time.Now()

		if b, err := eh.handler(args...); err != nil {
			log.Errorf("unable to execute event handler: %v", err)
		} else if !b {
			continue
		} else {
			log.WithField("handler", s).Debug("processed event handler")
		}

		s.Duration = time.Now().Sub(start).Milliseconds()
		result = append(result, s)
	}
	return result
}

func (sh *scriptHost) Every(every string, handler func(), opt common.JobOptions) (int, error) {
	do := func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Errorf("script error %v: %v", sh.Name(), r)
			}
		}()
		handler()
	}

	job, err := sh.engine.scheduler.Every(every, do, opt)

	if err != nil {
		return -1, err
	}

	id := len(sh.jobs)
	sh.jobs[id] = job

	return id, nil
}

func (sh *scriptHost) Cron(expr string, handler func(), opt common.JobOptions) (int, error) {
	do := func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Errorf("script error %v: %v", sh.Name(), r)
			}
		}()
		handler()
	}

	job, err := sh.engine.scheduler.Cron(expr, do, opt)
	if err != nil {
		return -1, err
	}

	id := len(sh.jobs)
	sh.jobs[id] = job

	return id, nil
}

func (sh *scriptHost) Cancel(jobId int) error {
	if j, ok := sh.jobs[jobId]; !ok {
		return fmt.Errorf("job not defined")
	} else {
		sh.engine.scheduler.Remove(j)
		return nil
	}
}

func (sh *scriptHost) On(event string, handler func(args ...interface{}) (bool, error), tags map[string]string) {
	h := &eventHandler{
		handler: handler,
		tags: map[string]string{
			"name":    sh.name,
			"file":    sh.name,
			"fileKey": sh.file.Info.Key(),
			"event":   event,
		},
	}

	for k, v := range tags {
		h.tags[k] = v
	}

	sh.events[event] = append(sh.events[event], h)
}

func (sh *scriptHost) close() {
	sh.Lock()
	defer sh.Unlock()

	if sh.jobs != nil {
		for _, j := range sh.jobs {
			sh.engine.scheduler.Remove(j)
		}
		sh.jobs = nil
	}

	if sh.script != nil {
		sh.script.Close()
	}

	for _, n := range sh.fakerNodes {
		err := n.Restore()
		if err != nil {
			log.Errorf("failed to close script properly: %v", err)
		}
	}
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

func (sh *scriptHost) Debug(args ...interface{}) {
	sh.engine.logger.Debug(args...)
}

func (sh *scriptHost) OpenFile(path string, hint string) (*dynamic.Config, error) {
	u, err := url.Parse(path)
	if err != nil || len(u.Scheme) == 0 || len(u.Opaque) > 0 {
		if !filepath.IsAbs(path) {
			if len(hint) > 0 {
				path = filepath.Join(hint, path)
			} else {
				p := getScriptPath(sh.file.Info.Kernel().Url)
				path = filepath.Join(filepath.Dir(p), path)
			}
		}

		u, err = file.ParseUrl(path)
		if err != nil {
			if len(hint) > 0 {
				return sh.OpenFile(path, "")
			}
			return nil, err
		}
	}

	f, err := sh.engine.reader.Read(u, nil)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (sh *scriptHost) KafkaClient() common.KafkaClient {
	return sh.engine.kafkaClient
}

func (sh *scriptHost) HttpClient() common.HttpClient {
	return &http.Client{Timeout: time.Second * 30}
}

func (sh *scriptHost) CanClose() bool {
	return len(sh.events) == 0 && len(sh.jobs) == 0 && len(sh.fakerNodes) == 0 && sh.script.CanClose()
}

func (sh *scriptHost) FindFakerTree(name string) common.FakerTree {
	t := generator.FindByName(name)
	ft := &fakerTree{
		t: t,
	}
	sh.fakerNodes = append(sh.fakerNodes, ft)
	return ft
}

func (sh *scriptHost) Lock() {
	sh.m.Lock()
}

func (sh *scriptHost) Unlock() {
	sh.m.Unlock()
}

func getScriptPath(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}
