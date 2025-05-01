package engine_test

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"mokapi/runtime/runtimetest"
	"testing"
	"time"
)

func TestEngine_Scheduler(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, e *engine.Engine, c *metrics.Counter)
	}{
		{
			name: "run job",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() { mokapi.sleep(1); });
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := events.GetEvents(events.NewTraits().WithNamespace("job").WithName("test.js"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(common.JobExecution)
				r.Equal(t, "test.js", exec.Tags["name"])
				r.Greater(t, exec.Duration, int64(0))
				r.Equal(t, "1s", exec.Schedule)

				r.Equal(t, float64(1), c.Value())
			},
		},
		{
			name: "run job with name",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {}, { tags: { name: 'foo' } });
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := events.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(common.JobExecution)
				r.Equal(t, "foo", exec.Tags["name"])
			},
		},
		{
			name: "run job with custom tag",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter) {
				err := e.AddScript(newScript("test.js", `
					import mokapi from 'mokapi'
					export default function() {
						mokapi.every('1s', function() {}, { tags: { foo: 'bar' } });
					}
				`))
				r.NoError(t, err)

				time.Sleep(300 * time.Millisecond)

				evts := events.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(common.JobExecution)
				r.Equal(t, "bar", exec.Tags["foo"])
			},
		},
		{
			name: "run job with logs",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter) {
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

				evts := events.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(common.JobExecution)
				r.Len(t, exec.Logs, 1)
				r.Equal(t, "a log message", exec.Logs[0].Message)
			},
		},
		{
			name: "run job script error",
			test: func(t *testing.T, e *engine.Engine, c *metrics.Counter) {
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

				evts := events.GetEvents(events.NewTraits().WithNamespace("job"))
				r.Len(t, evts, 1)
				exec := evts[0].Data.(common.JobExecution)
				r.NotNil(t, exec.Error)
				r.Equal(t, "Error: script error at test.js:5:13(3)", exec.Error.Message)
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

			events.SetStore(10, events.NewTraits().WithNamespace("job"))

			tc.test(t, e, app.Monitor.JobCounter)

			events.Reset()
		})
	}
}
