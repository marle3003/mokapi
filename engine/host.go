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
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type eventHandler struct {
	handler func(args ...interface{}) (bool, error)
	tags    map[string]string
}

type scriptHost struct {
	id       int
	name     string
	engine   *Engine
	script   common.Script
	jobs     map[int]*gocron.Job
	events   map[string][]*eventHandler
	cwd      string
	file     *config.Config
	checksum []byte
}

func newScriptHost(file *config.Config, e *Engine) *scriptHost {
	path := getScriptPath(file.Info.Url)

	sh := &scriptHost{
		id:       1,
		name:     file.Info.Path(),
		engine:   e,
		jobs:     make(map[int]*gocron.Job),
		events:   make(map[string][]*eventHandler),
		cwd:      filepath.Dir(path),
		file:     file,
		checksum: file.Checksum,
	}

	return sh
}

func (sh *scriptHost) Name() string {
	return sh.name
}

func (sh *scriptHost) Compile() error {
	s, err := sh.compile()
	if err != nil {
		return err
	}
	sh.script = s
	return nil
}

func (sh *scriptHost) Run() error {
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
	sh.engine.cron.Every(every)

	if opt.Times > 0 {
		sh.engine.cron.LimitRunsTo(opt.Times)
	}
	if !opt.RunFirstTimeImmediately {
		sh.engine.cron.WaitForSchedule()
	}

	j, err := sh.engine.cron.Do(func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Errorf("script error %v: %v", sh.Name(), r)
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

func (sh *scriptHost) Cron(expr string, handler func(), opt common.JobOptions) (int, error) {
	sh.engine.cron.Cron(expr)

	if opt.Times >= 0 {
		sh.engine.cron.LimitRunsTo(opt.Times)
	}
	if !opt.RunFirstTimeImmediately {
		sh.engine.cron.WaitForSchedule()
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
			"file":  sh.name,
			"event": event,
		},
	}

	for k, v := range tags {
		h.tags[k] = v
	}

	sh.events[event] = append(sh.events[event], h)
}

func (sh *scriptHost) close() {
	if sh.jobs != nil {
		for _, j := range sh.jobs {
			sh.engine.cron.RemoveByReference(j)
		}
		sh.jobs = nil
	}

	if sh.script != nil {
		sh.script.Close()
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

func (sh *scriptHost) OpenFile(path string, hint string) (*config.Config, error) {
	u, err := url.Parse(path)
	if err != nil || len(u.Scheme) == 0 {
		if !filepath.IsAbs(path) {
			if len(hint) > 0 {
				path = filepath.Join(hint, path)
			} else {
				path = filepath.Join(sh.cwd, path)
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

	f, err := sh.engine.reader.Read(u,
		config.WithParent(sh.file))
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
	return len(sh.events) == 0 && len(sh.jobs) == 0
}

func getScriptPath(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}

func (sh *scriptHost) compile() (common.Script, error) {
	s := sh.file.Data.(*script.Script)
	switch filepath.Ext(s.Filename) {
	case ".js":
		return js.New(getScriptPath(sh.file.Info.Url), s.Code, sh, sh.engine.jsConfig)
	case ".lua":
		return lua.New(getScriptPath(sh.file.Info.Url), s.Code, sh)
	default:
		return nil, fmt.Errorf("unsupported script %v", s.Filename)
	}
}