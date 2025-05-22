package engine

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/schema/json/generator"
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

	eventLogger func(level, message string)

	fakerNodes []*common.FakerTree
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
	if sh.engine.loader == nil {
		return fmt.Errorf("loader not defined")
	}
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
		sh.Lock()
		action := &common.Action{
			Tags: eh.tags,
		}
		sh.startEventHandler(action.AppendLog)
		start := time.Now()

		if b, err := eh.handler(args...); err != nil {
			log.Errorf("unable to execute event handler: %v", err)
			action.Error = &common.Error{Message: err.Error()}
		} else if !b {
			sh.Unlock()
			continue
		} else {
			log.WithField("handler", action).Debug("processed event handler")
		}

		action.Duration = time.Now().Sub(start).Milliseconds()
		action.Parameters = getDeepCopy(args)
		result = append(result, action)
		sh.endEventHandler()
		sh.Unlock()
	}
	return result
}

func (sh *scriptHost) Every(every string, handler func(), opt common.JobOptions) (int, error) {
	id := len(sh.jobs)

	do := sh.newJobFunc(handler, opt, every, id)
	job, err := sh.engine.scheduler.Every(every, do, opt)

	if err != nil {
		return -1, err
	}

	sh.jobs[id] = job

	return id, nil
}

func (sh *scriptHost) Cron(expr string, handler func(), opt common.JobOptions) (int, error) {
	id := len(sh.jobs)

	do := sh.newJobFunc(handler, opt, expr, id)
	job, err := sh.engine.scheduler.Cron(expr, do, opt)
	if err != nil {
		return -1, err
	}

	sh.jobs[id] = job

	return id, nil
}

func (sh *scriptHost) newJobFunc(handler func(), opt common.JobOptions, schedule string, id int) func() {
	tags := map[string]string{
		"name":    sh.name,
		"file":    sh.name,
		"fileKey": sh.file.Info.Key(),
	}
	for k, v := range opt.Tags {
		tags[k] = v
	}

	t := events.NewTraits().WithNamespace("job").WithName(tags["name"])
	if len(events.GetStores(t)) == 1 {
		events.SetStore(int(sh.engine.cfgEvent.Store["Default"].Size), t)
	}
	t = t.With("jobId", fmt.Sprintf("%d", id))
	counter := 1

	return func() {
		sh.Lock()
		defer sh.Unlock()
		job := sh.jobs[id]

		sh.engine.jobCounter.Add(1)

		exec := common.JobExecution{
			Schedule: schedule,
			MaxRuns:  opt.Times,
			Runs:     counter,
			NextRun:  job.NextRun(),
			Tags:     tags,
		}

		defer func() {
			r := recover()
			if r != nil {
				log.Errorf("script error %v: %v", sh.Name(), r)
				exec.Error = &common.Error{Message: fmt.Sprintf("%v", r)}
			}

			sh.endEventHandler()

			err := events.Push(exec, t)
			if err != nil {
				log.Errorf("failed to push event: %v", err)
			}
		}()

		sh.startEventHandler(exec.AppendLog)
		start := time.Now()

		handler()

		exec.Duration = time.Now().Sub(start).Milliseconds()
		counter++
	}
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
	if sh.eventLogger != nil && sh.IsLevelEnabled("info") {
		sh.eventLogger("log", fmt.Sprint(args...))
	}
}

func (sh *scriptHost) Warn(args ...interface{}) {
	sh.engine.logger.Warn(args...)
	if sh.eventLogger != nil && sh.IsLevelEnabled("warn") {
		sh.eventLogger("warn", fmt.Sprint(args...))
	}
}

func (sh *scriptHost) Error(args ...interface{}) {
	sh.engine.logger.Error(args...)
	if sh.eventLogger != nil && sh.IsLevelEnabled("error") {
		sh.eventLogger("error", fmt.Sprint(args...))
	}
}

func (sh *scriptHost) Debug(args ...interface{}) {
	sh.engine.logger.Debug(args...)
	if sh.eventLogger != nil && sh.IsLevelEnabled("debug") {
		sh.eventLogger("debug", fmt.Sprint(args...))
	}
}

func (sh *scriptHost) IsLevelEnabled(level string) bool {
	return sh.engine.IsLevelEnabled(level)
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

func (sh *scriptHost) HttpClient(opts common.HttpClientOptions) common.HttpClient {
	return &http.Client{
		Timeout: time.Second * 30,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if l := len(via); l > opts.MaxRedirects {
				log.Warnf("Stopped after %d redirects, original URL was %s", opts.MaxRedirects, via[0].URL)
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func (sh *scriptHost) CanClose() bool {
	return len(sh.events) == 0 && len(sh.jobs) == 0 && len(sh.fakerNodes) == 0 && sh.script.CanClose()
}

func (sh *scriptHost) FindFakerNode(name string) *common.FakerTree {
	t := generator.FindByName(name)
	if t == nil {
		return nil
	}
	ft := common.NewFakerTree(t)
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

func (sh *scriptHost) startEventHandler(eventLog func(level, message string)) {
	sh.eventLogger = eventLog
}

// Function to end the event handler context
func (sh *scriptHost) endEventHandler() {
	sh.eventLogger = nil
}

func getDeepCopy(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		b, _ := json.Marshal(arg)
		result[i] = string(b)
	}
	return result
}
