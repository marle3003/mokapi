package engine_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"testing"
)

func TestEventHandler(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		logger *enginetest.Logger
		run    func(evt common.EventEmitter) []*common.Action
		test   func(t *testing.T, actions []*common.Action, err error)
	}{
		{
			name: "script error",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', () => {
		throw new Error('script error')
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http")
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Equal(t, "Error: script error at test.js:4:8(3)", actions[0].Error.Message)
			},
		},
		{
			name: "console.log",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', () => {
		console.log('a log message from event handler')
	}, { track: true })
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http")
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Len(t, actions[0].Logs, 1)
				require.Equal(t, "a log message from event handler", actions[0].Logs[0].Message)
				require.Equal(t, "log", actions[0].Logs[0].Level)
			},
		},
		{
			name: "console.warn",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', () => {
		console.warn('a log message from event handler')
	}, { track: true })
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http")
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Len(t, actions[0].Logs, 1)
				require.Equal(t, "a log message from event handler", actions[0].Logs[0].Message)
				require.Equal(t, "warn", actions[0].Logs[0].Level)
			},
		},
		{
			name: "console.warn but not match log level",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', () => {
		console.warn('a log message from event handler')
	}, { track: true })
}
`,
			logger: &enginetest.Logger{IsLevelEnabledFunc: func(level string) bool { return false }},
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http")
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Len(t, actions[0].Logs, 0)
			},
		},
		{
			name: "parameter",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (msg) => {
	}, { track: true })
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http", "foo")
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Len(t, actions[0].Parameters, 1)
				require.Equal(t, `"foo"`, actions[0].Parameters[0])
			},
		},
		{
			name: "parameter should be a copy",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req) => {
		req.method = 'GET'
	})
	on('http', (req) => {
		req.method = 'DELETE'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				req := struct {
					Method string
				}{Method: "POST"}
				return evt.Emit("http", &req)
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 2)

				require.Len(t, actions[0].Parameters, 1)
				require.Equal(t, `{"Method":"GET"}`, actions[0].Parameters[0])

				require.Len(t, actions[1].Parameters, 1)
				require.Equal(t, `{"Method":"DELETE"}`, actions[1].Parameters[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []engine.Options
			if tc.logger != nil {
				opts = append(opts, engine.WithLogger(tc.logger))
			}

			e := enginetest.NewEngine(opts...)
			err := e.AddScript(newScript("test.js", tc.script))

			var actions []*common.Action
			if err == nil {
				actions = tc.run(e)
			}
			tc.test(t, actions, err)
		})
	}
}
