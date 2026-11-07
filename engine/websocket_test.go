package engine_test

import (
	"context"
	"io"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"net/http/httptest"
	"testing"
	"time"

	ws "github.com/coder/websocket"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestEvent_Websocket(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "emit connect",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				var called []string
				d.Register("foo", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketConnectEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "connect")
						return true, nil
					},
				})
				d.Register("bar", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketMessageEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "message")
						return true, nil
					},
				})

				d.EmitWebsocketConnect(&common.WebsocketConnectEvent{
					WebsocketEvent: common.WebsocketEvent{
						Type: common.WebsocketConnectEventType,
					},
				})

				require.Equal(t, []string{"connect"}, called)
			},
		},
		{
			name: "emit connect filtering API",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				called := map[string]bool{
					"foo": false,
					"bar": false,
				}
				for api := range called {
					d.Register("key", &engine.WebsocketEventHandler{
						Filter: common.WebsocketFilter{
							Api: api,
						},
						Execute: func(ctx *common.EventContext) (bool, error) {
							called[api] = true
							return true, nil
						},
					})
				}

				d.EmitWebsocketConnect(&common.WebsocketConnectEvent{
					WebsocketEvent: common.WebsocketEvent{
						Api: "foo",
					},
				})

				require.True(t, called["foo"], "foo")
				require.False(t, called["bar"], "bar")
			},
		},
		{
			name: "emit message",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				var called []string
				d.Register("foo", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketConnectEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "connect")
						return true, nil
					},
				})
				d.Register("bar", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketMessageEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "message")
						return true, nil
					},
				})

				d.EmitWebsocketMessage(&common.WebsocketMessageEvent{
					WebsocketEvent: common.WebsocketEvent{
						Type: common.WebsocketMessageEventType,
					},
				})

				require.Equal(t, []string{"message"}, called)
			},
		},
		{
			name: "emit message with channel filter",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				var called []string
				d.Register("foo", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type:    common.WebsocketMessageEventType,
						Channel: "/foo",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "foo")
						return true, nil
					},
				})
				d.Register("bar", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type:    common.WebsocketMessageEventType,
						Channel: "/bar",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "bar")
						return true, nil
					},
				})

				d.EmitWebsocketMessage(&common.WebsocketMessageEvent{
					WebsocketEvent: common.WebsocketEvent{
						Type: common.WebsocketMessageEventType,
						Channel: &common.WebsocketEventChannel{
							Key: "/foo",
						},
					},
				})

				require.Equal(t, []string{"foo"}, called)
			},
		},
		{
			name: "emit message with channel template filter",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				var called []string
				d.Register("foo", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type:    common.WebsocketMessageEventType,
						Channel: "/chat/{roomId}",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "/chat/{roomId}")
						return true, nil
					},
				})
				d.Register("/bar", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type:    common.WebsocketMessageEventType,
						Channel: "/chat/general",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "/chat/general")
						return true, nil
					},
				})

				d.EmitWebsocketMessage(&common.WebsocketMessageEvent{
					WebsocketEvent: common.WebsocketEvent{
						Type: common.WebsocketMessageEventType,
						Channel: &common.WebsocketEventChannel{
							Key: "/chat/{roomId}",
						},
					},
				})

				require.Equal(t, []string{"/chat/{roomId}"}, called)
			},
		},
		{
			name: "emit close",
			test: func(t *testing.T) {
				d := &engine.WebsocketEventDispatcher{}

				var called []string
				d.Register("foo", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketConnectEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "connect")
						return true, nil
					},
				})
				d.Register("bar", &engine.WebsocketEventHandler{
					Filter: common.WebsocketFilter{
						Type: common.WebsocketCloseEventType,
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "close")
						return true, nil
					},
				})

				d.EmitWebsocketClose(&common.WebsocketCloseEvent{
					WebsocketEvent: common.WebsocketEvent{
						Type: common.WebsocketCloseEventType,
					},
				})

				require.Equal(t, []string{"close"}, called)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestWebsocketClient(t *testing.T) {
	logrus.SetOutput(io.Discard)

	createCfg := func(topic string, msg *asyncapi3.Message) *asyncapi3.Config {
		ch := asyncapi3test.NewChannel(
			asyncapi3test.UseMessage("foo",
				&asyncapi3.MessageRef{Value: msg},
			),
		)

		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel(topic, ch),
			asyncapi3test.WithComponentMessage("foo", msg),
			asyncapi3test.WithOperation("sendAction",
				asyncapi3test.WithOperationAction("send"),
				asyncapi3test.WithOperationChannel(ch),
				asyncapi3test.UseOperationMessage(msg),
			),
		)
	}

	testcases := []struct {
		name string
		cfg  func() *asyncapi3.Config
		test func(t *testing.T, e *engine.Engine, app *runtime.App, url string)
	}{
		{
			name: "send simple",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")))
				return createCfg("/foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App, url string) {
				hook := test.NewGlobal()

				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/foo", &ws.DialOptions{})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = e.AddScript(newScript("test.js", `
					import { app } from 'mokapi'
					export default function() {
						const result = app.websocket().send('/foo', 'hello', {
							result: (result) => console.log(result)
						})
					}
				`))
				require.NoError(t, err)
				require.Equal(t, `{"api":"foo","channel":"/foo","message":"hello"}`, hook.LastEntry().Message)

				_, b, err := c.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, `"hello"`, string(b))
			},
		},
		{
			name: "send no message",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")))
				return createCfg("/foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App, url string) {
				hook := test.NewGlobal()

				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/foo", &ws.DialOptions{})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = e.AddScript(newScript("test.js", `
					import { app } from 'mokapi'
					export default function() {
						const result = app.websocket().send('/foo', undefined, {
							result: (result) => console.log(result)
						})
					}
				`))
				require.NoError(t, err)
				require.Equal(t, `{"api":"foo","channel":"/foo","message":"fnsy"}`, hook.LastEntry().Message)

				_, b, err := c.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, `"fnsy"`, string(b))
			},
		},
		{
			name: "send empty channel",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")))
				return createCfg("/foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App, url string) {
				hook := test.NewGlobal()

				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/foo", &ws.DialOptions{})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = e.AddScript(newScript("test.js", `
					import { app } from 'mokapi'
					export default function() {
						const result = app.websocket().send('', undefined, {
							result: (result) => console.log(result)
						})
					}
				`))
				require.NoError(t, err)
				require.Equal(t, `{"api":"foo","channel":"/foo","message":"fnsy"}`, hook.LastEntry().Message)

				_, b, err := c.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, `"fnsy"`, string(b))
			},
		},
		{
			name: "send with api but channel not exists",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithPayload(schematest.New("string")))
				return createCfg("/foo", msg)
			},
			test: func(t *testing.T, e *engine.Engine, app *runtime.App, url string) {
				err := e.AddScript(newScript("test.js", `
					import { app } from 'mokapi'
					export default function() {
						const result = app.websocket().api('bar').send('/foo', undefined, {
							result: (result) => console.log(result),
							retry: { retries: 0 }
						})
					}
				`))
				require.EqualError(t, err, "websocket API 'bar' not found at reflect.methodValueCall (native)")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			e := enginetest.NewEngine(
				engine.WithWebsocketClient(engine.NewWebsocketClient(app)),
				engine.WithDefaultLogger(),
			)

			cfg := tc.cfg()
			s, err := app.Websocket.Add(getConfig(cfg), e)
			require.NoError(t, err)
			serv := httptest.NewServer(s)
			defer serv.Close()

			tc.test(t, e, app, serv.URL)
		})
	}
}
