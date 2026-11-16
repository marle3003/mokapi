package engine

import (
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"net/url"
	"runtime"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

type hostTestLogger struct{}

func (*hostTestLogger) Info(...interface{})        {}
func (*hostTestLogger) Warn(...interface{})        {}
func (*hostTestLogger) Error(...interface{})       {}
func (*hostTestLogger) Debug(...interface{})       {}
func (*hostTestLogger) IsLevelEnabled(string) bool { return true }

func TestScriptHost_EventLogger(t *testing.T) {
	tests := []struct {
		name     string
		act      func(sh *scriptHost)
		expected []hostTestLog
	}{
		{
			name: "log before event handler ends",
			act: func(sh *scriptHost) {
				sh.Info("message")
				sh.endEventHandler()
			},
			expected: []hostTestLog{{level: "log", message: "message"}},
		},
		{
			name: "log after event handler ends",
			act: func(sh *scriptHost) {
				sh.endEventHandler()
				sh.Info("message")
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs []hostTestLog
			sh := &scriptHost{engine: &Engine{logger: &hostTestLogger{}}}
			sh.startEventHandler(func(level, message string) {
				logs = append(logs, hostTestLog{level: level, message: message})
			})

			test.act(sh)

			r.Equal(t, test.expected, logs)
		})
	}
}

func TestScriptHost_EndEventHandlerWaitsForActiveLogger(t *testing.T) {
	sh := &scriptHost{engine: &Engine{logger: &hostTestLogger{}}}
	loggerStarted := make(chan struct{})
	releaseLogger := make(chan struct{})
	loggerDone := make(chan struct{})
	handlerEnded := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-releaseLogger:
		default:
			close(releaseLogger)
		}
	})

	var logs []hostTestLog
	// Keep the logger active until endEventHandler is waiting for it.
	sh.startEventHandler(func(level, message string) {
		close(loggerStarted)
		<-releaseLogger
		logs = append(logs, hostTestLog{level: level, message: message})
	})

	go func() {
		sh.Info("message")
		close(loggerDone)
	}()
	<-loggerStarted

	go func() {
		sh.endEventHandler()
		close(handlerEnded)
	}()

	// A pending writer makes TryRLock fail, proving that endEventHandler
	// is waiting instead of clearing the logger while it is in use.
	writerWaiting := false
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for !writerWaiting {
		select {
		case <-handlerEnded:
			r.FailNow(t, "event handler ended while logger was active")
		case <-deadline.C:
			r.FailNow(t, "event handler did not attempt to acquire the logger lock")
		default:
		}

		if sh.eventLoggerMu.TryRLock() {
			sh.eventLoggerMu.RUnlock()
			runtime.Gosched()
			continue
		}
		writerWaiting = true
	}

	// Once the logger returns, both operations must finish and retain the log.
	close(releaseLogger)
	select {
	case <-loggerDone:
	case <-deadline.C:
		r.FailNow(t, "logger did not finish")
	}
	select {
	case <-handlerEnded:
	case <-deadline.C:
		r.FailNow(t, "event handler did not finish")
	}

	r.Equal(t, []hostTestLog{{level: "log", message: "message"}}, logs)
}

type hostTestLog struct {
	level   string
	message string
}

type stubJob struct {
	next time.Time
}

func (j stubJob) NextRun() time.Time { return j.next }

type stubScheduler struct {
	removed []Job
}

func (*stubScheduler) Start()                                               {}
func (*stubScheduler) Close()                                               {}
func (*stubScheduler) Every(string, func(), common.JobOptions) (Job, error) { return nil, nil }
func (*stubScheduler) Cron(string, func(), common.JobOptions) (Job, error)  { return nil, nil }
func (s *stubScheduler) Remove(job Job)                                     { s.removed = append(s.removed, job) }

func newTestHost(jobs map[int]*jobEntry) *scriptHost {
	return &scriptHost{
		name: "/scripts/test.js",
		file: &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "file", Path: "/scripts/test.js"}}},
		jobs: jobs,
	}
}

func newTestEntry(maxRuns int, job Job) *jobEntry {
	return &jobEntry{
		job:      job,
		schedule: "5m",
		maxRuns:  maxRuns,
		tags:     map[string]string{"name": "test.js"},
		state:    newJobState(maxRuns),
	}
}

func TestScriptHost_Jobs(t *testing.T) {
	t.Run("job lookup after close finds nothing", func(t *testing.T) {
		sh := &scriptHost{jobs: map[int]*jobEntry{0: newTestEntry(0, stubJob{})}}
		r.NotNil(t, sh.entry(0))

		sh.takeJobs()

		r.Nil(t, sh.entry(0))
	})

	t.Run("cancel unknown job", func(t *testing.T) {
		sh := &scriptHost{jobs: map[int]*jobEntry{}}

		r.EqualError(t, sh.Cancel(1), "job not defined")
	})

	t.Run("jobs registered", func(t *testing.T) {
		sh := &scriptHost{jobs: map[int]*jobEntry{0: newTestEntry(0, stubJob{})}}
		r.True(t, sh.hasJobs())

		sh.takeJobs()

		r.False(t, sh.hasJobs())
	})

	t.Run("cancel removes the job from the scheduler and marks it cancelled", func(t *testing.T) {
		scheduler := &stubScheduler{}
		job := stubJob{next: time.Now().Add(time.Hour)}
		sh := newTestHost(map[int]*jobEntry{0: newTestEntry(0, job)})
		sh.engine = &Engine{scheduler: scheduler}

		r.NoError(t, sh.Cancel(0))

		r.Equal(t, []Job{job}, scheduler.removed)
		info := sh.listJobs()
		r.Len(t, info, 1)
		r.Equal(t, common.JobStateCancelled, info[0].State)
		r.True(t, info[0].NextRun.IsZero())
	})

	t.Run("cancel a job without scheduler job", func(t *testing.T) {
		scheduler := &stubScheduler{}
		sh := newTestHost(map[int]*jobEntry{0: newTestEntry(1, nil)})
		sh.engine = &Engine{scheduler: scheduler}

		r.NoError(t, sh.Cancel(0))

		r.Equal(t, []Job{nil}, scheduler.removed)
		r.Equal(t, common.JobStateCancelled, sh.listJobs()[0].State)
	})
}

func TestJobState(t *testing.T) {
	run := common.JobRun{EventId: "evt", Status: common.JobStatusSuccess}

	tests := []struct {
		name    string
		maxRuns int
		act     func(s *jobState)
		runs    int
		state   string
		hasRun  bool
	}{
		{name: "not run yet", maxRuns: -1, act: func(s *jobState) {},
			runs: 0, state: common.JobStateScheduled},
		{name: "one run", maxRuns: -1, act: func(s *jobState) { s.finishRun(run) },
			runs: 1, state: common.JobStateScheduled, hasRun: true},
		{name: "unlimited job never finishes", maxRuns: -1,
			act:  func(s *jobState) { s.finishRun(run); s.finishRun(run); s.finishRun(run) },
			runs: 3, state: common.JobStateScheduled, hasRun: true},
		{name: "limited job is scheduled until its last run", maxRuns: 2,
			act:  func(s *jobState) { s.finishRun(run) },
			runs: 1, state: common.JobStateScheduled, hasRun: true},
		{name: "limited job finishes with its last run", maxRuns: 2,
			act:  func(s *jobState) { s.finishRun(run); s.finishRun(run) },
			runs: 2, state: common.JobStateFinished, hasRun: true},
		{name: "limited job with one run finishes after it", maxRuns: 1,
			act:  func(s *jobState) { s.finishRun(run) },
			runs: 1, state: common.JobStateFinished, hasRun: true},
		{name: "cancelled job", maxRuns: -1, act: func(s *jobState) { s.cancel() },
			runs: 0, state: common.JobStateCancelled},
		{name: "cancelled wins over a finishing run", maxRuns: 1,
			act:  func(s *jobState) { s.cancel(); s.finishRun(run) },
			runs: 1, state: common.JobStateCancelled, hasRun: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newJobState(tc.maxRuns)

			tc.act(s)

			snapshot := s.snapshot()
			r.Equal(t, tc.runs, snapshot.runs)
			r.Equal(t, tc.state, snapshot.state)
			if tc.hasRun {
				r.Equal(t, &run, snapshot.lastRun)
			} else {
				r.Nil(t, snapshot.lastRun)
			}
		})
	}

	t.Run("next run number counts the run in progress", func(t *testing.T) {
		s := newJobState(-1)
		r.Equal(t, 1, s.nextRunNumber())
		s.finishRun(run)
		r.Equal(t, 2, s.nextRunNumber())
	})
}

func TestScriptHost_ListJobs(t *testing.T) {
	newHost := newTestHost

	t.Run("registered job that has not run", func(t *testing.T) {
		next := time.Now().Add(time.Hour)
		sh := newHost(map[int]*jobEntry{0: newTestEntry(3, stubJob{next: next})})

		info := sh.listJobs()

		r.Len(t, info, 1)
		r.Equal(t, sh.file.Info.Key()+"-0", info[0].Id)
		r.Equal(t, "test.js", info[0].Name)
		r.Equal(t, "5m", info[0].Schedule)
		r.Equal(t, 3, info[0].MaxRuns)
		r.Equal(t, 0, info[0].Runs)
		r.Equal(t, common.JobStateScheduled, info[0].State)
		r.Equal(t, next, info[0].NextRun)
		r.Equal(t, map[string]string{"name": "test.js"}, info[0].Tags)
		r.Nil(t, info[0].LastRun)
	})

	t.Run("job after a run", func(t *testing.T) {
		entry := newTestEntry(-1, stubJob{})
		run := common.JobRun{EventId: "evt", Time: time.Now(), Duration: 12, Status: common.JobStatusWarning}
		entry.state.finishRun(run)
		sh := newHost(map[int]*jobEntry{0: entry})

		info := sh.listJobs()

		r.Equal(t, 1, info[0].Runs)
		r.Equal(t, &run, info[0].LastRun)
		r.True(t, info[0].NextRun.IsZero())
	})

	t.Run("finished job has no next run", func(t *testing.T) {
		entry := newTestEntry(1, stubJob{next: time.Now().Add(time.Hour)})
		entry.state.finishRun(common.JobRun{})
		sh := newHost(map[int]*jobEntry{0: entry})

		info := sh.listJobs()

		r.Equal(t, common.JobStateFinished, info[0].State)
		r.True(t, info[0].NextRun.IsZero())
	})

	t.Run("finished job without scheduler job has no next run", func(t *testing.T) {
		entry := newTestEntry(1, nil)
		entry.state.finishRun(common.JobRun{})
		sh := newHost(map[int]*jobEntry{0: entry})

		info := sh.listJobs()

		r.Equal(t, common.JobStateFinished, info[0].State)
		r.True(t, info[0].NextRun.IsZero())
	})

	t.Run("id uses the file key even if the fileKey tag is overridden", func(t *testing.T) {
		entry := newTestEntry(-1, stubJob{})
		entry.tags["fileKey"] = "custom"
		sh := newHost(map[int]*jobEntry{0: entry})

		info := sh.listJobs()

		r.Equal(t, sh.file.Info.Key()+"-0", info[0].Id)
		r.Equal(t, "custom", info[0].Tags["fileKey"])
	})

	t.Run("jobs are ordered by their index", func(t *testing.T) {
		sh := newHost(map[int]*jobEntry{
			2: newTestEntry(-1, stubJob{}),
			0: newTestEntry(-1, stubJob{}),
			1: newTestEntry(-1, stubJob{}),
		})

		info := sh.listJobs()

		key := sh.file.Info.Key()
		r.Equal(t, []string{key + "-0", key + "-1", key + "-2"}, []string{info[0].Id, info[1].Id, info[2].Id})
	})

	t.Run("no jobs after close", func(t *testing.T) {
		sh := newHost(map[int]*jobEntry{0: newTestEntry(-1, stubJob{})})

		sh.takeJobs()

		r.Empty(t, sh.listJobs())
	})
}
