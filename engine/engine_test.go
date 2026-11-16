package engine_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"mokapi/runtime/runtimetest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	r "github.com/stretchr/testify/require"
)

func TestEngine_Scheduler(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager)
	}{
		{
			name: "run job",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() { mokapi.sleep(1); });
					}
				`))
				r.NoError(t, err)

				time.Sleep(time.Millisecond * 800)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("job").WithName("test.js"))
				r.Len(t, evts, 1)
				r.Equal(t, "namespace=job, name=test.js, jobId=0", evts[0].Traits.String())
				exec := evts[0].Data.(*common.JobExecution)
				r.Equal(t, "test.js", exec.Tags["name"])
				r.Greater(t, exec.Duration, int64(0))
				r.Equal(t, "1s", exec.Schedule)
				r.Equal(t, -1, exec.MaxRuns)
				r.Equal(t, 1, exec.Runs)
				r.False(t, exec.NextRun.IsZero())
				r.True(t, exec.NextRun.After(evts[0].Time))

				r.Equal(t, float64(1), c.Value())
			},
		},
		{
			name: "run cron job immediately does not block (deadlock guard for the inline first run)",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				done := make(chan error)
				go func() {
					done <- e.AddScript(newScript("test.js", `
						import mokapi from 'mokapi'
						export default function() {
							mokapi.cron('0 0 1 1 *', function() {}, { skipImmediateFirstRun: false });
						}
					`))
				}()

				select {
				case err := <-done:
					r.NoError(t, err)
				case <-time.After(5 * time.Second):
					t.Fatal("registering a cron job with an immediate first run blocks")
				}

				r.Eventually(t, func() bool {
					return len(sm.GetEvents(events.NewTraits().WithNamespace("job").WithName("test.js"))) == 1
				}, 2*time.Second, 10*time.Millisecond)
			},
		},
		{
			name: "run job with name",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {}, { tags: { name: 'foo' } });
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(*common.JobExecution)
				r.Equal(t, "foo", exec.Tags["name"])
			},
		},
		{
			name: "run job with custom tag",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {}, { tags: { foo: 'bar' } });
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(*common.JobExecution)
				r.Equal(t, "bar", exec.Tags["foo"])
			},
		},
		{
			name: "run job with logs",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {
							console.log('a log message');
						});
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(*common.JobExecution)
				r.Len(t, exec.Logs, 1)
				r.Equal(t, "a log message", exec.Logs[0].Message)
			},
		},
		{
			name: "run job script error",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				logrus.SetLevel(logrus.DebugLevel)
				t.Cleanup(func() { logrus.SetLevel(logrus.InfoLevel) })
				hook := test.NewGlobal()
				t.Cleanup(hook.Reset)

				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {
							throw new Error('script error');
						});
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := sm.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(*common.JobExecution)
				r.NotNil(t, exec.Error)
				r.Equal(t, "Error: script error at test.js:5:13(3)", exec.Error.Message)
				r.True(t, hasLogEntry(hook.Entries, logrus.DebugLevel, "goroutine "), "expected panic stack trace in debug logs")
			},
		},
		{
			name: "unsupported script should not return error",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter, sm *events.StoreManager) {
				err := e.AddScript(newScript("test.png", ""))
				r.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtimetest.NewApp()
			app.Monitor.JobCounter = &metrics.Counter{}
			e := enginetest.NewEngine(
				engine.WithScheduler(engine.NewDefaultScheduler()),
				engine.WithApp(app),
			)
			defer e.Close()
			e.Start()

			tc.test(t, e, app.Monitor.JobCounter, app.Events)
		})
	}
}

func TestEngine_AddScript_Delete(t *testing.T) {
	logrus.SetOutput(io.Discard)

	jobScript := `
		import mokapi from 'mokapi'
		export default function() {
			mokapi.every('100ms', function() {});
		}
	`
	jobEvents := func(sm *events.StoreManager) int {
		return len(sm.GetEvents(events.NewTraits().WithNamespace("job").WithName("test.js")))
	}
	deleted := func(evt dynamic.ConfigEvent) dynamic.ConfigEvent {
		evt.Event = dynamic.Delete
		return evt
	}

	testcases := []struct {
		name string
		test func(t *testing.T, e *engine.Engine, sm *events.StoreManager)
	}{
		{
			name: "deleted script is removed and does not run again",
			test: func(t *testing.T, e *engine.Engine, sm *events.StoreManager) {
				script := newScript("test.js", jobScript)
				r.NoError(t, e.AddScript(script))
				r.Equal(t, 1, e.Scripts())
				r.Eventually(t, func() bool { return jobEvents(sm) > 0 }, 2*time.Second, 10*time.Millisecond)

				r.NoError(t, e.AddScript(deleted(script)))

				r.Equal(t, 0, e.Scripts())
				// an execution already in flight at delete time may still be recorded once
				count := jobEvents(sm)
				r.Never(t, func() bool { return jobEvents(sm) > count+1 }, 400*time.Millisecond, 10*time.Millisecond,
					"job of deleted script must not fire anymore")
			},
		},
		{
			name: "delete of unknown script does nothing",
			test: func(t *testing.T, e *engine.Engine, sm *events.StoreManager) {
				other := newScript("other.js", jobScript)
				r.NoError(t, e.AddScript(other))

				r.NoError(t, e.AddScript(deleted(newScript("test.js", jobScript))))

				r.Equal(t, 1, e.Scripts())
			},
		},
		{
			name: "delete of script that already closed itself",
			test: func(t *testing.T, e *engine.Engine, sm *events.StoreManager) {
				script := newScript("test.js", `export default function() {}`)
				r.NoError(t, e.AddScript(script))
				r.Equal(t, 0, e.Scripts())

				r.NoError(t, e.AddScript(deleted(script)))

				r.Equal(t, 0, e.Scripts())
			},
		},
		{
			name: "script added again after delete runs again",
			test: func(t *testing.T, e *engine.Engine, sm *events.StoreManager) {
				script := newScript("test.js", jobScript)
				r.NoError(t, e.AddScript(script))
				r.NoError(t, e.AddScript(deleted(script)))
				r.Equal(t, 0, e.Scripts())

				r.NoError(t, e.AddScript(script))

				r.Equal(t, 1, e.Scripts())
				r.Eventually(t, func() bool { return jobEvents(sm) > 0 }, 2*time.Second, 10*time.Millisecond)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtimetest.NewApp()
			e := enginetest.NewEngine(
				engine.WithScheduler(engine.NewDefaultScheduler()),
				engine.WithApp(app),
			)
			defer e.Close()
			e.Start()

			tc.test(t, e, app.Events)
		})
	}
}

func TestEngine_Jobs(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// jobs with a skipped first run keep the order tests free of timing
	idleJobs := func(tags string, count int) string {
		var calls strings.Builder
		for i := 0; i < count; i++ {
			calls.WriteString("mokapi.every('1h', function() {}, { skipImmediateFirstRun: true" + tags + " });\n")
		}
		return "import mokapi from 'mokapi'\nexport default function() {\n" + calls.String() + "}"
	}
	deleted := func(evt dynamic.ConfigEvent) dynamic.ConfigEvent {
		evt.Event = dynamic.Delete
		return evt
	}
	describe := func(jobs []common.JobInfo) []string {
		var result []string
		for _, j := range jobs {
			result = append(result, j.Name+" "+j.Tags["file"]+" "+j.Id[strings.LastIndex(j.Id, "-")+1:])
		}
		return result
	}

	testcases := []struct {
		name string
		test func(t *testing.T, e *engine.Engine)
	}{
		{
			name: "no scripts",
			test: func(t *testing.T, e *engine.Engine) {
				r.Empty(t, e.Jobs())
			},
		},
		{
			name: "sorted by name, then file, then job index",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("b.js", idleJobs("", 2))))
				r.NoError(t, e.AddScript(newScript("dir2/same.js", idleJobs(", tags: { name: 'a.js' }", 1))))
				r.NoError(t, e.AddScript(newScript("a.js", idleJobs("", 1))))
				r.NoError(t, e.AddScript(newScript("dir1/same.js", idleJobs(", tags: { name: 'a.js' }", 2))))

				r.Equal(t, []string{
					"a.js a.js 0",
					"a.js dir1/same.js 0",
					"a.js dir1/same.js 1",
					"a.js dir2/same.js 0",
					"b.js b.js 0",
					"b.js b.js 1",
				}, describe(e.Jobs()))
			},
		},
		{
			name: "job that has not run is listed as scheduled",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", idleJobs(", times: 3", 1))))

				jobs := e.Jobs()

				r.Len(t, jobs, 1)
				r.Equal(t, "1h", jobs[0].Schedule)
				r.Equal(t, 3, jobs[0].MaxRuns)
				r.Equal(t, 0, jobs[0].Runs)
				r.Equal(t, common.JobStateScheduled, jobs[0].State)
				r.False(t, jobs[0].NextRun.IsZero())
				r.Nil(t, jobs[0].LastRun)
			},
		},
		{
			name: "limited job finishes without next run",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('100ms', function() {}, { times: 2 });
					}
				`)))

				r.Eventually(t, func() bool { return e.Jobs()[0].State == common.JobStateFinished }, 2*time.Second, 10*time.Millisecond)
				job := e.Jobs()[0]
				r.Equal(t, 2, job.Runs)
				r.True(t, job.NextRun.IsZero())
				r.NotNil(t, job.LastRun)
			},
		},
		{
			name: "last run refers to the stored event and has the run status",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1h', function() { console.warn('careful'); });
					}
				`)))

				r.Eventually(t, func() bool { return e.Jobs()[0].LastRun != nil }, 2*time.Second, 10*time.Millisecond)
				job := e.Jobs()[0]
				r.Equal(t, 1, job.Runs)
				r.Equal(t, common.JobStatusWarning, job.LastRun.Status)
				r.NotEmpty(t, job.LastRun.EventId)
				r.False(t, job.LastRun.Time.IsZero())
			},
		},
		{
			name: "failed run is counted",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1h', function() { throw new Error('boom'); });
					}
				`)))

				r.Eventually(t, func() bool { return e.Jobs()[0].LastRun != nil }, 2*time.Second, 10*time.Millisecond)
				job := e.Jobs()[0]
				r.Equal(t, 1, job.Runs)
				r.Equal(t, common.JobStatusError, job.LastRun.Status)
			},
		},
		{
			name: "reloaded script lists only its new jobs",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", idleJobs("", 2))))
				r.Len(t, e.Jobs(), 2)

				r.NoError(t, e.AddScript(newScript("test.js", idleJobs(", tags: { name: 'reloaded' }", 1))))

				jobs := e.Jobs()
				r.Len(t, jobs, 1)
				r.Equal(t, "reloaded", jobs[0].Name)
			},
		},
		{
			name: "jobs of a deleted script disappear",
			test: func(t *testing.T, e *engine.Engine) {
				script := newScript("test.js", idleJobs("", 1))
				r.NoError(t, e.AddScript(script))
				r.NoError(t, e.AddScript(newScript("other.js", idleJobs("", 1))))
				r.Len(t, e.Jobs(), 2)

				r.NoError(t, e.AddScript(deleted(script)))

				jobs := e.Jobs()
				r.Len(t, jobs, 1)
				r.Equal(t, "other.js", jobs[0].Name)
			},
		},
		{
			name: "jobs can be listed while they run",
			test: func(t *testing.T, e *engine.Engine) {
				r.NoError(t, e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('10ms', function() {});
						mokapi.every('10ms', function() { throw new Error('boom'); });
					}
				`)))

				deadline := time.Now().Add(300 * time.Millisecond)
				for time.Now().Before(deadline) {
					for _, j := range e.Jobs() {
						_ = j.Runs
					}
				}

				for _, j := range e.Jobs() {
					r.Greater(t, j.Runs, 1)
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtimetest.NewApp()
			e := enginetest.NewEngine(
				engine.WithScheduler(engine.NewDefaultScheduler()),
				engine.WithApp(app),
			)
			defer e.Close()
			e.Start()

			tc.test(t, e)
		})
	}
}

func hasLogEntry(entries []logrus.Entry, level logrus.Level, message string) bool {
	for _, entry := range entries {
		if entry.Level == level && strings.Contains(entry.Message, message) {
			return true
		}
	}
	return false
}
