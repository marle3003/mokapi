package mokapi_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/mokapi"
	"mokapi/js/require"
	"testing"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestModule_On(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "register event handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var event string
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					event = evt
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					let result = 0;
					m.on('http', () => result++)
				`)
				r.NoError(t, err)
				r.Equal(t, "http", event)
				b, err := handler(&common.EventContext{})
				r.NoError(t, err)
				r.Equal(t, false, b)
				v, _ := vm.RunString("result")
				r.Equal(t, int64(1), v.Export())
			},
		},
		{
			name: "event handler with parameter",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					let result = false
					m.on('http', (param) => result = param === 'foo')
				`)
				r.NoError(t, err)
				b, err := handler(&common.EventContext{Args: []any{"foo"}})
				r.NoError(t, err)
				r.Equal(t, true, b)
				v, _ := vm.RunString("result")
				r.Equal(t, true, v.Export())
			},
		},
		{
			name: "event handler changes params",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', (param) => { param['foo'] = false })
				`)
				r.NoError(t, err)
				b, err := handler(&common.EventContext{Args: []any{map[string]bool{"foo": true}}})
				r.NoError(t, err)
				r.Equal(t, true, b)
			},
		},
		{
			name: "event handler does not change params",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', (param) => { })
				`)
				r.NoError(t, err)
				b, err := handler(&common.EventContext{Args: []any{map[string]bool{"foo": true}}})
				r.NoError(t, err)
				r.Equal(t, false, b)
			},
		},
		{
			name: "event handler does not change params but uses track argument",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', (param) => { }, { track: true })
				`)
				r.NoError(t, err)
				b, err := handler(&common.EventContext{Args: []any{map[string]bool{"foo": true}}})
				r.NoError(t, err)
				r.Equal(t, true, b)
			},
		},
		{
			name: "event handler changes params but disables track",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', (param) => { param['foo'] = false }, { track: false })
				`)
				r.NoError(t, err)
				b, err := handler(&common.EventContext{Args: []any{map[string]bool{"foo": true}}})
				r.NoError(t, err)
				r.Equal(t, false, b)
			},
		},
		{
			name: "event handler throws error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => { throw new Error('TEST') })
				`)
				r.NoError(t, err)
				_, err = handler(&common.EventContext{})
				r.EqualError(t, err, "Error: TEST at <eval>:3:33(3)")
			},
		},
		{
			name: "event handler with tags",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var tags map[string]string
				host.OnFunc = func(evt string, do common.EventHandler, t map[string]string) {
					tags = t
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, { tags: { foo: 'bar', bar: null } })
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]string{"foo": "bar", "bar": "null"}, tags)
			},
		},
		{
			name: "event handler with tags but invalid type",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, { tags: 'foo' })
				`)
				r.EqualError(t, err, "unexpected type for tags: String at mokapi/js/mokapi.(*Module).On-fm (native)")
			},
		},
		{
			name: "event handler invalid type for args",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', () => true, 'foo')
				`)
				r.EqualError(t, err, "unexpected type for args: String at mokapi/js/mokapi.(*Module).On-fm (native)")
			},
		},
		{
			name: "async event handler",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				var handler common.EventHandler
				host.OnFunc = func(evt string, do common.EventHandler, tags map[string]string) {
					handler = do
				}

				_, err := vm.RunString(`
					const m = require('mokapi')
					m.on('http', async (p) => {
						p.msg = await getMessage();
					})

					let getMessage = async () => {
						return new Promise(async (resolve, reject) => {
						  setTimeout(() => {
							resolve('foo');
						  }, 200);
						});
					}
				`)
				r.NoError(t, err)
				p := &struct {
					Msg string `json:"msg"`
				}{}
				_, err = handler(&common.EventContext{Args: []any{p}})
				r.NoError(t, err)
				r.Equal(t, "foo", p.Msg)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg, err := require.NewRegistry()
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

func TestModule_On_Run(t *testing.T) {
	testcases := []struct {
		name   string
		script string
		logger *enginetest.Logger
		run    func(evt common.EventEmitter) []*common.Action
		test   func(t *testing.T, actions []*common.Action, err error)
	}{
		{
			name: "response header using CanonicalHeaderKey",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.headers['content-type'] = 'text/plain'
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{
					Headers: map[string]any{"Content-Type": "application/json"},
				}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Equal(t, "text/plain", res.Headers["Content-Type"])
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.Nil(t, actions[0].Error)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Len(t, res.Headers, 1)
				r.Equal(t, "text/plain", res.Headers["Content-Type"])
			},
		},
		{
			name: "set response data",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.data = { "foo": "bar" }
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Equal(t, &map[string]interface{}{"foo": "bar"}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.Nil(t, actions[0].Error)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, map[string]interface{}{"foo": "bar"}, res.Data)
			},
		},
		{
			name: "set status code",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.statusCode = 201
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Equal(t, 201, res.StatusCode)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.Nil(t, actions[0].Error)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, 201, res.StatusCode)
			},
		},
		{
			name: "set status code but wrong type",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.statusCode = 'foo'
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http", &common.EventRequest{}, &common.EventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.NotNil(t, actions[0].Error)
				r.Equal(t, actions[0].Error.Message, "statusCode must be a Integer at <eval>:4:6(3)")

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, 0, res.StatusCode)
			},
		},
		{
			name: "set body",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.body = 'hello world'
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Equal(t, "hello world", res.Body)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.Nil(t, actions[0].Error)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, "hello world", res.Body)
			},
		},
		{
			name: "set object to body",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.body = { foo: 'bar' }
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				return evt.Emit("http", &common.EventRequest{}, &common.EventResponse{})
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.NotNil(t, actions[0].Error)
				r.Equal(t, "body must be a String at <eval>:4:6(5)", actions[0].Error.Message)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, "", res.Body)
			},
		},
		{
			name: "set array to data and push item",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.data = [ 1, 2 ]
	res.data.push(3)
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Equal(t, &[]any{int64(1), int64(2), int64(3)}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)
				r.Nil(t, actions[0].Error)
				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, []any{float64(1), float64(2), float64(3)}, res.Data)
			},
		},
		{
			name: "set object and change field",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.data = { foo: "bar" }
	res.data.foo = 'yuh'
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Nil(t, actions[0].Error)
				r.Equal(t, &map[string]any{"foo": "yuh"}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, map[string]any{"foo": "yuh"}, res.Data)
			},
		},
		{
			name: "change field on given data",
			script: `
const m = require('mokapi')
m.on('http', (req, res) => {
	res.data.foo = 'yuh'
})
`,
			run: func(evt common.EventEmitter) []*common.Action {
				res := &common.EventResponse{Data: map[string]any{"foo": "bar"}}
				actions := evt.Emit("http", &common.EventRequest{}, res)
				r.Nil(t, actions[0].Error)
				r.Equal(t, map[string]any{"foo": "yuh"}, res.Data)
				return actions
			},
			test: func(t *testing.T, actions []*common.Action, err error) {
				r.NoError(t, err)

				var res *common.EventResponse
				err = json.Unmarshal([]byte(actions[0].Parameters[1].(string)), &res)
				r.Equal(t, map[string]any{"foo": "yuh"}, res.Data)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := require.NewRegistry()
			reg.RegisterNativeModule("mokapi", mokapi.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			var runEvent common.EventHandler
			host.OnFunc = func(event string, do common.EventHandler, tags map[string]string) {
				runEvent = do
			}

			_, err = vm.RunString(tc.script)
			r.NoError(t, err)

			var actions []*common.Action
			e := enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*common.Action {
				ctx := &common.EventContext{
					Args: args,
				}
				_, err := runEvent(ctx)
				a := &common.Action{}
				for _, arg := range args {
					b, _ := json.Marshal(arg)
					a.Parameters = append(a.Parameters, string(b))
				}
				if err != nil {
					a.Error = &common.Error{Message: err.Error()}
				}
				actions = append(actions, a)
				return actions
			})
			tc.run(e)

			tc.test(t, actions, err)
		})
	}
}
