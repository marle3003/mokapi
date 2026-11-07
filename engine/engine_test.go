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

func hasLogEntry(entries []logrus.Entry, level logrus.Level, message string) bool {
	for _, entry := range entries {
		if entry.Level == level && strings.Contains(entry.Message, message) {
			return true
		}
	}
	return false
}
