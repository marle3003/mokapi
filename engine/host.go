package engine

import (
	"fmt"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/dynamic/script"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/lua"
	"net/url"
	"path/filepath"
	"time"
)

type eventHandler struct {
	handler func(args ...interface{}) (bool, error)
	tags    map[string]string
}

type scriptHost struct {
	id     int
	Name   string
	engine *Engine
	script common.Script
	jobs   map[int]*gocron.Job
	events map[string][]*eventHandler
	cwd    string
	file   *config.Config
}

func newScriptHost(file *config.Config, e *Engine) (*scriptHost, error) {
	path := getScriptName(file.Url)

	sh := &scriptHost{
		id:     1,
		Name:   path,
		engine: e,
		jobs:   make(map[int]*gocron.Job),
		events: make(map[string][]*eventHandler),
		cwd:    filepath.Dir(path),
		file:   file,
	}

	src := file.Data.(*script.Script).Code

	var err error
	switch filepath.Ext(path) {
	case ".js":
		if sh.script, err = js.New(path, src, sh); err != nil {
			return nil, err
		}
	case ".lua":
		if sh.script, err = lua.New(path, src, sh); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported script %v", path)
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
		} else {
			log.Debugf("processed event handler %v", s)
		}

		s.Duration = time.Now().Sub(start)
		result = append(result, s)
	}
	return result
}

func (sh *scriptHost) Every(every string, handler func(), times int, tags map[string]string) (int, error) {
	sh.engine.cron.Every(every)

	if times > 0 {
		sh.engine.cron.LimitRunsTo(times)
	}

	j, err := sh.engine.cron.Do(func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Errorf("script error %v: %v", sh.Name, r)
			}
		}()
		handler()
	})
	if err != nil {
		return -1, err
	}

	id := len(sh.jobs)
	sh.jobs[id] = j

	return id, nil
}

func (sh *scriptHost) Cron(expr string, handler func(), times int, tags map[string]string) (int, error) {
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
			"name":  sh.Name,
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

	// todo: read as plaintext
	f, err := sh.engine.reader.Read(u, config.AsPlaintext())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", f.Data), nil
}

func (sh *scriptHost) OpenScript(path string) (common.Script, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(sh.cwd, path)
	}

	u, err := file.ParseUrl(path)
	if err != nil {
		return nil, err
	}

	name := getScriptName(u)
	if s, ok := sh.engine.scripts[name]; ok {
		return s.script, nil
	}

	file, err := sh.engine.reader.Read(u, config.WithParent(sh.file))
	if err != nil {
		return nil, err
	}

	file.Listeners = append(file.Listeners, func(c *config.Config) {
		if old, ok := sh.engine.scripts[sh.Name]; ok && c.Version <= old.file.Version {
			return
		}
		log.Infof("remove file: %v", c.Url)
		sh.engine.remove(name)
	})

	err = sh.engine.AddScript(file)
	if err != nil {
		return nil, err
	}
	return sh.engine.scripts[name].script, nil
}

func (sh *scriptHost) KafkaClient() common.KafkaClient {
	return sh.engine.kafkaClient
}

func getScriptName(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}
