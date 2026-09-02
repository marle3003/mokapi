package engine_test

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js/mokapi"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestEventHandler_Http(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		logger *enginetest.Logger
		run    func(evt common.EventEmitter) []*common.Action
		test   func(t *testing.T, actions []*common.Action, hook *test.Hook, err error)
	}{
		{
			name: "response header using CanonicalHeaderKey",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.headers['content-type'] = 'text/plain'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{
					Headers: map[string]any{"Content-Type": "application/json"},
				}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Equal(t, "text/plain", mokapi.Export(res.Headers["Content-Type"]))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, _ *test.Hook, err error) {
				require.NoError(t, err)
				require.Nil(t, actions[0].Error)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Len(t, res.Headers, 1)
				require.Equal(t, "text/plain", res.Headers["Content-Type"])
			},
		},
		{
			name: "set response data",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data = { "foo": "bar" }
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Equal(t, &map[string]interface{}{"foo": "bar"}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, _ *test.Hook, err error) {
				require.NoError(t, err)
				require.Nil(t, actions[0].Error)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, res.Data)
			},
		},
		{
			name: "set status code",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.statusCode = 201
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Equal(t, 201, res.StatusCode)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, _ *test.Hook, err error) {
				require.NoError(t, err)
				require.Nil(t, actions[0].Error)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, 201, res.StatusCode)
			},
		},
		{
			name: "set status code but wrong type",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.statusCode = 'foo'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)
				require.NotNil(t, actions[0].Error)
				require.Equal(t, "failed to set statusCode: expected Integer but got String at test.js:4:6(3)", actions[0].Error.Message)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, 0, res.StatusCode)
				require.Len(t, hook.Entries, 2)
				require.Equal(t, "unable to execute event handler: failed to set statusCode: expected Integer but got String at test.js:4:6(3)", hook.LastEntry().Message)
			},
		},
		{
			name: "set body",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.body = 'hello world'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Equal(t, "hello world", res.Body)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, _ *test.Hook, err error) {
				require.NoError(t, err)
				require.Nil(t, actions[0].Error)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, "hello world", res.Body)
			},
		},
		{
			name: "set object to body",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.body = { foo: 'bar' }
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.EmitHttp(&common.HttpEventRequest{}, &common.HttpEventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)
				require.NotNil(t, actions[0].Error)
				require.Equal(t, "failed to set body: expected String but got Object at test.js:4:6(5)", actions[0].Error.Message)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, "", res.Body)
				require.Len(t, hook.Entries, 2)
				require.Equal(t, "unable to execute event handler: failed to set body: expected String but got Object at test.js:4:6(5)", hook.LastEntry().Message)
			},
		},
		{
			name: "set array to data and push item",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data = [ 1, 2 ]
		res.data.push(3)
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Equal(t, &[]any{int64(1), int64(2), int64(3)}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)
				require.Nil(t, actions[0].Error)
				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, []any{float64(1), float64(2), float64(3)}, res.Data)
			},
		},
		{
			name: "set object and change field",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data = { foo: "bar" }
		res.data.foo = 'yuh'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Nil(t, actions[0].Error)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
			},
		},
		{
			name: "change field on given data",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data.foo = 'yuh'
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Data: map[string]any{"foo": "bar"}}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Nil(t, actions[0].Error)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
			},
		},
		{
			name: "using Object.assign",
			script: `import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data = Object.assign(res.data, {foo: 'yuh'})
	})
}
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Data: map[string]any{"foo": "bar"}}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Nil(t, actions[0].Error)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, hook *test.Hook, err error) {
				require.NoError(t, err)

				var res *common.HttpEventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "yuh"}, mokapi.Export(res.Data))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			hook := test.NewGlobal()
			logrus.SetLevel(logrus.InfoLevel)

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
			tc.test(t, actions, hook, err)
		})
	}
}

func TestEventHandler_Priority(t *testing.T) {
	testcases := []struct {
		name    string
		scripts []string
		run     func(evt common.EventEmitter) []*common.Action
		test    func(t *testing.T, actions []*common.Action)
	}{
		{
			name: "handlers in same script",
			scripts: []string{`import { on } from 'mokapi'
export default () => {
    let counter = 0
	on('http', (req, res) => {
		res.data.foo = 'handler1';
		res.data.handler1 = counter++ 
	}, { priority: -1 })
	on('http', (req, res) => {
		res.data.foo = 'handler2';
		res.data.handler2 = counter++ 
	}, { priority: 10 })
	on('http', (req, res) => {
		res.data.foo = 'handler3';
		res.data.handler3 = counter++ 
	})
}
`,
			},
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Data: map[string]any{"foo": "bar"}}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Nil(t, actions[0].Error)
				require.Nil(t, actions[1].Error)
				require.Nil(t, actions[2].Error)
				require.Equal(t, map[string]any{"foo": "handler1", "handler1": int64(2), "handler2": int64(0), "handler3": int64(1)}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action) {
				var res *common.HttpEventResponse
				_ = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler2", "handler2": float64(0)}, mokapi.Export(res.Data))
				_ = json.Unmarshal([]byte(actions[1].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler3", "handler2": float64(0), "handler3": float64(1)}, mokapi.Export(res.Data))
				_ = json.Unmarshal([]byte(actions[2].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler1", "handler2": float64(0), "handler3": float64(1), "handler1": float64(2)}, mokapi.Export(res.Data))
			},
		},
		{
			name: "handlers in different scripts",
			scripts: []string{`
import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data.foo = 'handler1';
	}, { priority: -1 })
}
`,
				`
import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data.foo = 'handler2';
	}, { priority: 10 })
}
`,
				`
import { on } from 'mokapi'
export default () => {
	on('http', (req, res) => {
		res.data.foo = 'handler3';
	})
}
`,
			},
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Data: map[string]any{"foo": "bar"}}
				actions := evt.EmitHttp(&common.HttpEventRequest{}, res)
				require.Nil(t, actions[0].Error)
				require.Nil(t, actions[1].Error)
				require.Nil(t, actions[2].Error)
				require.Equal(t, map[string]any{"foo": "handler1"}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action) {
				var res *common.HttpEventResponse
				_ = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler2"}, mokapi.Export(res.Data))
				_ = json.Unmarshal([]byte(actions[1].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler3"}, mokapi.Export(res.Data))
				_ = json.Unmarshal([]byte(actions[2].Parameters[1]), &res)
				require.Equal(t, map[string]any{"foo": "handler1"}, mokapi.Export(res.Data))
			},
		},
		{
			name: "using context and shared number",
			scripts: []string{`
import { app, shared } from 'mokapi'
export default () => {
	const db = shared.update('db', (current) => current || {});
	db['foo'] = 123;
	app.http()
	  .use((req, res) => { res.context.data = db['foo'] })
	  .get('/foo', (req, res) => { res.data = res.context.data })
}
`,
			},
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Context: make(map[string]any)}
				actions := evt.EmitHttp(&common.HttpEventRequest{Method: http.MethodGet, Key: "/foo"}, res)
				require.Equal(t, int64(123), mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action) {
				var res *common.HttpEventResponse
				// first action is from use handler
				_ = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Nil(t, res.Data)
				// second action is from use handler
				_ = json.Unmarshal([]byte(actions[1].Parameters[1]), &res)
				require.Equal(t, float64(123), mokapi.Export(res.Data))
			},
		},
		{
			name: "using context and shared object",
			scripts: []string{`
import { app, shared } from 'mokapi'
export default () => {
	const db = shared.update('db', (current) => current || {});
	db['foo'] = { value: 123 };
	app.http()
	  .use((req, res) => { res.context.data = db['foo'] })
	  .get('/foo', (req, res) => { res.data = res.context.data })
}
`,
			},
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.HttpEventResponse{Context: make(map[string]any)}
				actions := evt.EmitHttp(&common.HttpEventRequest{Method: http.MethodGet, Key: "/foo"}, res)
				require.Equal(t, map[string]any{"value": int64(123)}, mokapi.Export(res.Data))
				return actions
			},
			test: func(t *testing.T, actions []*common.Action) {
				var res *common.HttpEventResponse
				// first action is from use handler
				_ = json.Unmarshal([]byte(actions[0].Parameters[1]), &res)
				require.Nil(t, res.Data)
				// second action is from use handler
				_ = json.Unmarshal([]byte(actions[1].Parameters[1]), &res)
				require.Equal(t, map[string]any{"value": float64(123)}, mokapi.Export(res.Data))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)

			var opts []engine.Options
			e := enginetest.NewEngine(opts...)
			for i, s := range tc.scripts {
				err := e.AddScript(newScript(fmt.Sprintf("test-%v.js", i), s))
				require.NoError(t, err)
			}

			actions := tc.run(e)
			tc.test(t, actions)
		})
	}
}
