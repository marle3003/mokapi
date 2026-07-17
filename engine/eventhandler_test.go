package engine_test

import (
	"io"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestEventHandler(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name   string
		script string
		logger *enginetest.Logger
		run    func(evt common.HttpEventEmitter) []*common.Action
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
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
			name: "calling console.log tracks the event handler",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', () => {
		console.log('a log message from event handler')
	})
}
`,
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 1)
				require.Len(t, actions[0].Parameters, 2)
				require.Equal(t, `{"method":"","url":{"scheme":"","host":"","port":0,"path":"","query":""},"body":null,"path":null,"query":null,"header":null,"cookie":null,"querystring":null,"api":"","key":"","operationId":""}`, actions[0].Parameters[0])
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
			run: func(evt common.HttpEventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{Method: http.MethodPost}, &common.HttpEventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				require.NoError(t, err)
				require.Len(t, actions, 2)

				require.Len(t, actions[0].Parameters, 2)
				require.Contains(t, actions[0].Parameters[0], `"method":"GET"`)

				require.Len(t, actions[1].Parameters, 2)
				require.Contains(t, actions[1].Parameters[0], `"method":"DELETE"`)
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
