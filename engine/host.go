package engine

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"maps"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/schema/json/generator"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Snapshotter interface {
	Snapshot() any
}

type scriptHost struct {
	name string

	engine *Engine

	script common.Script
	jobs   map[int]*jobEntry
	jobsMu sync.Mutex // guards jobs; separate from m, which close() holds while closing the script
	file   *dynamic.Config

	eventLogger   func(level, message string)
	eventLoggerMu sync.RWMutex
	cleanupFuncs  []func()
	sm            *events.StoreManager

	m sync.Mutex
}

func newScriptHost(file *dynamic.Config, engine *Engine) *scriptHost {
	sh := &scriptHost{
		name:   file.Info.Path(),
		jobs:   make(map[int]*jobEntry),
		file:   file,
		engine: engine,
	}
	if engine.app != nil {
		sh.sm = engine.app.Events
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
	log.Infof("executing script %v", sh.name)
	return sh.script.Run()
}

// Every and Cron are only called from the script's event loop (single caller). That keeps the
// id allocation in Cron safe without holding jobsMu across scheduler.Cron, whose inline
// immediate run would otherwise deadlock or invert the lock order with the scheduler's mutex.
func (sh *scriptHost) Every(every string, handler func(), opt common.JobOptions) (int, error) {
	// The scheduler may run the job immediately. Holding the lock until the job is
	// registered makes the first run wait, so it can look up its own job.
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()

	id := len(sh.jobs)
	entry := sh.newJobEntry(every, opt)
	do := sh.newJobFunc(handler, entry, id)
	job, err := sh.engine.scheduler.Every(every, do, opt)
	if err != nil {
		return -1, err
	}

	entry.job = job
	sh.jobs[id] = entry
	return id, nil
}

func (sh *scriptHost) Cron(expr string, handler func(), opt common.JobOptions) (int, error) {
	id := sh.nextJobId()

	entry := sh.newJobEntry(expr, opt)
	do := sh.newJobFunc(handler, entry, id)
	job, err := sh.engine.scheduler.Cron(expr, do, opt)
	if err != nil {
		return -1, err
	}

	// must not hold jobsMu across scheduler.Cron: an immediate first run executes inline
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	entry.job = job
	sh.jobs[id] = entry
	return id, nil
}

func (sh *scriptHost) nextJobId() int {
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	return len(sh.jobs)
}

// entry returns nil if the job is not registered (yet) or the host is closed.
func (sh *scriptHost) entry(id int) *jobEntry {
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	return sh.jobs[id]
}

func (sh *scriptHost) newJobEntry(schedule string, opt common.JobOptions) *jobEntry {
	tags := map[string]string{
		"name":    filepath.Base(sh.name),
		"file":    sh.name,
		"fileKey": sh.file.Info.Key(),
	}
	maps.Copy(tags, opt.Tags)

	return &jobEntry{
		schedule: schedule,
		maxRuns:  opt.Times,
		tags:     tags,
		state:    newJobState(opt.Times),
	}
}

func (sh *scriptHost) newJobFunc(handler func(), entry *jobEntry, id int) func() {
	t := events.NewTraits().WithNamespace("job").WithName(entry.tags["name"])
	if len(sh.sm.GetStores(t)) == 1 {
		sh.sm.SetStore(int(sh.engine.cfgEvent.Store["Default"].Size), t)
	}
	t = t.With("jobId", fmt.Sprintf("%d", id))

	return func() {
		var nextRun time.Time
		if e := sh.entry(id); e != nil && e.job != nil {
			nextRun = e.job.NextRun()
		} else {
			log.Debugf("no job found with id %d", id)
		}

		sh.engine.app.Monitor.JobCounter.Add(1)

		exec := common.JobExecution{
			Schedule: entry.schedule,
			MaxRuns:  entry.maxRuns,
			Runs:     entry.state.nextRunNumber(),
			NextRun:  nextRun,
			Tags:     entry.tags,
		}

		start := time.Now()

		defer func() {
			r := recover()
			if r != nil {
				log.Debugf("script error %v: %v", sh.Name(), string(debug.Stack()))
				log.Errorf("script error %v: %v", sh.Name(), r)
				exec.Error = &common.Error{Message: fmt.Sprintf("%v", r)}
			}

			sh.endEventHandler()

			evt, err := sh.sm.PushEvent(&exec, t)
			if err != nil {
				log.Errorf("failed to push event: %v", err)
			}
			entry.state.finishRun(common.JobRun{
				EventId:  evt.Id,
				Time:     start,
				Duration: exec.Duration,
				Status:   exec.Status(),
			})
		}()

		sh.startEventHandler(exec.AppendLog)

		handler()

		exec.Duration = time.Now().Sub(start).Milliseconds()
	}
}

func (sh *scriptHost) Cancel(jobId int) error {
	entry := sh.entry(jobId)
	if entry == nil {
		return fmt.Errorf("job not defined")
	}
	sh.engine.scheduler.Remove(entry.job)
	entry.state.cancel()
	return nil
}

func (sh *scriptHost) close() {
	sh.Lock()
	defer sh.Unlock()

	for _, entry := range sh.takeJobs() {
		sh.engine.scheduler.Remove(entry.job)
	}

	if sh.script != nil {
		sh.script.Close()
	}

	for index := len(sh.cleanupFuncs) - 1; index >= 0; index-- {
		sh.cleanupFuncs[index]()
	}

	sh.engine.Clear(sh.name)
}

// takeJobs empties the job registry and returns the jobs that were registered.
func (sh *scriptHost) takeJobs() map[int]*jobEntry {
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	jobs := sh.jobs
	sh.jobs = nil
	return jobs
}

func (sh *scriptHost) hasJobs() bool {
	sh.jobsMu.Lock()
	defer sh.jobsMu.Unlock()
	return len(sh.jobs) > 0
}

func (sh *scriptHost) AddCleanupFunc(f func()) {
	sh.cleanupFuncs = append(sh.cleanupFuncs, f)
}

func (sh *scriptHost) Info(args ...any) {
	sh.engine.logger.Info(args...)
	sh.logEvent("info", "log", args...)
}

func (sh *scriptHost) Warn(args ...any) {
	sh.engine.logger.Warn(args...)
	sh.logEvent("warn", "warn", args...)
}

func (sh *scriptHost) Error(args ...any) {
	sh.engine.logger.Error(args...)
	sh.logEvent("error", "error", args...)
}

func (sh *scriptHost) Debug(args ...any) {
	sh.engine.logger.Debug(args...)
	sh.logEvent("debug", "debug", args...)
}

func (sh *scriptHost) SetEventLogger(logger func(level, message string)) {
	sh.eventLoggerMu.Lock()
	defer sh.eventLoggerMu.Unlock()
	sh.eventLogger = logger
}

func (sh *scriptHost) logEvent(configuredLevel, eventLevel string, args ...any) {
	if !sh.IsLevelEnabled(configuredLevel) {
		return
	}

	sh.eventLoggerMu.RLock()
	defer sh.eventLoggerMu.RUnlock()
	if sh.eventLogger != nil {
		sh.eventLogger(eventLevel, fmt.Sprint(args...))
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
				cwd := sh.Cwd()
				path = filepath.Join(cwd, path)
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

func (sh *scriptHost) MqttClient() common.MqttClient {
	return sh.engine.mqttClient
}

func (sh *scriptHost) HttpClient(opts common.HttpClientOptions) common.HttpClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if opts.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if l := len(via); l > opts.MaxRedirects {
				log.Warnf("Stopped after %d redirects, original URL was %s", opts.MaxRedirects, via[0].URL)
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return c
}

func (sh *scriptHost) WebsocketClient() common.WebsocketClient {
	return sh.engine.websocketClient
}

func (sh *scriptHost) CanClose() bool {
	if sh.script == nil {
		return true
	}
	return !sh.engine.Has(sh.name) && !sh.hasJobs() && sh.script.CanClose() && len(sh.cleanupFuncs) == 0
}

func (sh *scriptHost) FindFakerNode(name string) *generator.Node {
	return generator.FindByName(name)
}

func (sh *scriptHost) Lock() {
	sh.m.Lock()
}

func (sh *scriptHost) Unlock() {
	sh.m.Unlock()
}

func (sh *scriptHost) Store() common.Store {
	return sh.engine.store
}

func (sh *scriptHost) Cwd() string {
	u := sh.file.Info.Kernel().Url
	if len(u.Path) > 0 {
		return filepath.Dir(u.Path)
	}
	return filepath.Dir(u.Opaque)
}

func (sh *scriptHost) Webhook(name string, url string, args common.WebhookArgs) (*common.WebhookResponse, error) {
	wh, err := sh.engine.app.Http.Webhook(name, args)
	if err != nil {
		return nil, err
	}
	res, err := wh.RunWebhook(url, args)
	if err != nil {
		return nil, fmt.Errorf("webhook %s failed: %w", name, err)
	}
	return res, nil
}

func getScriptPath(u *url.URL) string {
	if len(u.Path) > 0 {
		return u.Path
	}
	return u.Opaque
}

func (sh *scriptHost) startEventHandler(eventLog func(level, message string)) {
	sh.SetEventLogger(eventLog)
}

// Function to end the event handler context
func (sh *scriptHost) endEventHandler() {
	sh.SetEventLogger(nil)
}

func getDeepCopy(args []any) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		if s, ok := arg.(Snapshotter); ok {
			arg = s.Snapshot()
		}
		b, _ := json.Marshal(arg)
		result[i] = string(b)
	}
	return result
}
