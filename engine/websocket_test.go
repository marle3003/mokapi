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
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
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

func TestWebsocketClient_Send_MessageId(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// channel "/dm" with the messages directMessage and ack, channel "/ping" with a message without payload
	newConfig := func() *asyncapi3.Config {
		direct := asyncapi3test.NewMessage(
			asyncapi3test.WithPayload(schematest.New("object",
				schematest.WithProperty("kind", schematest.New("string", schematest.WithEnumValues("direct"))),
				schematest.WithRequired("kind"),
			)),
		)
		ack := asyncapi3test.NewMessage(
			asyncapi3test.WithPayload(schematest.New("object",
				schematest.WithProperty("kind", schematest.New("string", schematest.WithEnumValues("ack"))),
				schematest.WithRequired("kind"),
			)),
		)

		dm := asyncapi3test.NewChannel(
			asyncapi3test.UseMessage("directMessage", &asyncapi3.MessageRef{Value: direct}),
			asyncapi3test.UseMessage("ack", &asyncapi3.MessageRef{Value: ack}),
		)
		ping := asyncapi3test.NewChannel(
			asyncapi3test.UseMessage("ping", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
		)
		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("/dm", dm),
			asyncapi3test.AddChannel("/ping", ping),
		)
	}

	testcases := []struct {
		name    string
		channel string
		args    common.WebsocketSendArgs
		// received is empty if the call failed
		test func(t *testing.T, received string, err error)
	}{
		{
			name: "message is generated from the selected message",
			args: common.WebsocketSendArgs{MessageId: "ack"},
			test: func(t *testing.T, received string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kind":"ack"}`, received)
			},
		},
		{
			name: "valid message for the selected message is sent as given",
			args: common.WebsocketSendArgs{MessageId: "ack", Message: map[string]any{"kind": "ack", "extra": 1}},
			test: func(t *testing.T, received string, err error) {
				require.NoError(t, err)
				require.JSONEq(t, `{"kind":"ack","extra":1}`, received)
			},
		},
		{
			name: "message is validated against the selected message",
			// valid for ack, but directMessage is selected
			args: common.WebsocketSendArgs{MessageId: "directMessage", Message: map[string]any{"kind": "ack"}},
			test: func(t *testing.T, received string, err error) {
				require.Error(t, err)
				require.Empty(t, received)
			},
		},
		{
			name: "unknown messageId is an error without fallback",
			args: common.WebsocketSendArgs{MessageId: "akc"},
			test: func(t *testing.T, received string, err error) {
				require.EqualError(t, err, "message 'akc' not found in channel '/dm'")
				require.Empty(t, received)
			},
		},
		{
			name:    "selected message without payload sends no payload",
			channel: "/ping",
			args:    common.WebsocketSendArgs{MessageId: "ping"},
			test: func(t *testing.T, received string, err error) {
				require.NoError(t, err)
				// the message is null; the channel validates against payload schemas only
				// and writes an empty frame if it has none
				require.Empty(t, received)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			e := enginetest.NewEngine(engine.WithWebsocketClient(engine.NewWebsocketClient(app)))
			s, err := app.Websocket.Add(getConfig(newConfig()), e)
			require.NoError(t, err)
			serv := httptest.NewServer(s)
			defer serv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			channel := tc.channel
			if channel == "" {
				channel = "/dm"
			}
			c, _, err := ws.Dial(ctx, serv.URL+channel, &ws.DialOptions{})
			require.NoError(t, err)
			defer func() { _ = c.CloseNow() }()

			args := tc.args
			args.ChannelName = channel
			_, err = engine.NewWebsocketClient(app).Send(args)

			received := ""
			if err == nil {
				_, b, readErr := c.Read(ctx)
				require.NoError(t, readErr)
				received = string(b)
			}
			tc.test(t, received, err)
		})
	}
}

func TestWebsocketClient_Send_ScriptFile(t *testing.T) {
	logrus.SetOutput(io.Discard)

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	e := enginetest.NewEngine(engine.WithWebsocketClient(engine.NewWebsocketClient(app)))
	cfg := asyncapi3test.NewConfig(
		asyncapi3test.WithInfo("foo", "", ""),
		asyncapi3test.AddChannel("/chat", asyncapi3test.NewChannel(
			asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
		)),
	)
	s, err := app.Websocket.Add(getConfig(cfg), e)
	require.NoError(t, err)
	serv := httptest.NewServer(s)
	defer serv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := ws.Dial(ctx, serv.URL+"/chat", &ws.DialOptions{})
	require.NoError(t, err)
	defer func() { _ = c.CloseNow() }()

	_, err = engine.NewWebsocketClient(app).Send(common.WebsocketSendArgs{
		ChannelName: "/chat",
		Message:     map[string]any{"text": "hi"},
		ScriptFile:  "file:///scripts/sender.ts",
	})
	require.NoError(t, err)

	evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket").With("type", "message"))
	require.Len(t, evts, 1)
	require.Equal(t, "file:///scripts/sender.ts", evts[0].Data.(*websocket.MessageLog).Script)
}

func TestWebsocketClient_Send_UsesConfigAfterReload(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newCfg := func(messageId string) *asyncapi3.Config {
		msg := asyncapi3test.NewMessage(
			asyncapi3test.WithContentType("application/json"),
			asyncapi3test.WithPayload(schematest.New("object")),
		)
		return asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("/chat", asyncapi3test.NewChannel(
				asyncapi3test.UseMessage(messageId, &asyncapi3.MessageRef{Value: msg}),
			)),
		)
	}
	em := runtimetest.NewApp()
	st := websocket.New(newCfg("old"), enginetest.NewEngine(), em.Events, monitor.NewWebsocket())
	app := runtimetest.NewApp(runtimetest.WithWebsocketInfo("foo", runtime.NewWebsocketInfo(newCfg("new"), st)))
	st.Update(newCfg("new"))
	serv := httptest.NewServer(st)
	defer serv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := ws.Dial(ctx, serv.URL+"/chat", &ws.DialOptions{})
	require.NoError(t, err)
	defer func() { _ = c.CloseNow() }()

	_, err = engine.NewWebsocketClient(app).Send(common.WebsocketSendArgs{
		ChannelName: "/chat",
		Message:     map[string]any{"text": "hi"},
	})
	require.NoError(t, err)

	evts := em.Events.GetEvents(events.NewTraits().WithNamespace("websocket").With("type", "message"))
	require.Len(t, evts, 1)
	require.Equal(t, "new", evts[0].Data.(*websocket.MessageLog).MessageId)
}

func TestWebsocketClient_Send_ConfigRemovedWhileScriptIsRunning(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// an API whose last file was removed keeps its store but has no config snapshot (nil)
	newApp := func() *runtime.App {
		cfg := asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("foo", "", ""),
			asyncapi3test.AddChannel("/x", asyncapi3test.NewChannel(
				asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
			)),
		)
		st := websocket.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewWebsocket())
		return runtimetest.NewApp(runtimetest.WithWebsocketInfo("foo", runtime.NewWebsocketInfo(nil, st)))
	}

	testcases := []struct {
		name string
		args common.WebsocketSendArgs
		err  string
	}{
		{name: "API given", args: common.WebsocketSendArgs{Api: "foo", ChannelName: "/x"}, err: "websocket API 'foo' not found"},
		{name: "channel only skips it", args: common.WebsocketSendArgs{ChannelName: "/x"}, err: "websocket channel '/x' not found"},
		{name: "neither API nor channel", args: common.WebsocketSendArgs{}, err: "websocket channel '' not found"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := engine.NewWebsocketClient(newApp()).Send(tc.args)

				require.EqualError(t, err, tc.err)
			})
		})
	}
}

func TestWebsocketClient_Send_NoApiOrChannel(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		app  func() *runtime.App
	}{
		{
			name: "no API loaded",
			app:  func() *runtime.App { return runtimetest.NewApp() },
		},
		{
			name: "API without channels",
			app: func() *runtime.App {
				return runtimetest.NewApp(runtimetest.WithWebsocket(asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", ""))))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := engine.NewWebsocketClient(tc.app()).Send(common.WebsocketSendArgs{})

				require.EqualError(t, err, "websocket channel '' not found")
			})
		})
	}

	t.Run("not found is retried", func(t *testing.T) {
		app := runtimetest.NewApp()
		// set after the app is created, adding an API changes the log level
		logrus.SetLevel(logrus.DebugLevel)
		hook := test.NewGlobal()

		_, err := engine.NewWebsocketClient(app).Send(common.WebsocketSendArgs{
			Retry: common.RetryArgs{Retries: 2, InitialRetryTime: time.Millisecond, MaxRetryTime: time.Second, Factor: 2},
		})

		require.EqualError(t, err, "websocket channel '' not found")
		require.Contains(t, getMessages(hook), "websocket channel '' not found. Retry in 1ms")
	})
}

func TestWebsocketClient_Send_Api(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newConfig := func(name string, channels ...string) *asyncapi3.Config {
		opts := []asyncapi3test.ConfigOptions{asyncapi3test.WithInfo(name, "", "")}
		for _, name := range channels {
			opts = append(opts, asyncapi3test.AddChannel(name, asyncapi3test.NewChannel(
				asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
			)))
		}
		return asyncapi3test.NewConfig(opts...)
	}
	message := map[string]any{"text": "hi"}

	testcases := []struct {
		name string
		app  *runtime.App
		args common.WebsocketSendArgs
		// channel is the channel of the result, "/chat" if empty
		channel string
		err     string
	}{
		{
			name: "neither API nor channel",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"))),
			args: common.WebsocketSendArgs{},
		},
		{
			name: "channel only",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"))),
			args: common.WebsocketSendArgs{ChannelName: "/chat"},
		},
		{
			name: "API given",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"))),
			args: common.WebsocketSendArgs{Api: "foo", ChannelName: "/chat"},
		},
		{
			name: "two APIs without API and channel",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"), newConfig("bar", "/chat"))),
			args: common.WebsocketSendArgs{},
			err:  "ambiguous websocket API: specify the cluster",
		},
		{
			name: "two channels without API and channel",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat", "/news"))),
			args: common.WebsocketSendArgs{},
			err:  "ambiguous channel . Specify the API",
		},
		{
			name: "same channel in two APIs",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"), newConfig("bar", "/chat"))),
			args: common.WebsocketSendArgs{ChannelName: "/chat"},
			err:  "ambiguous channel /chat. Specify the cluster",
		},
		{
			name: "channel matching an address with parameters",
			app: runtimetest.NewApp(runtimetest.WithWebsocket(asyncapi3test.NewConfig(
				asyncapi3test.WithInfo("foo", "", ""),
				asyncapi3test.AddChannel("/rooms/{id}", asyncapi3test.NewChannel(
					asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
					asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: asyncapi3test.NewMessage()}),
				)),
			))),
			args:    common.WebsocketSendArgs{ChannelName: "/rooms/1"},
			channel: "/rooms/1",
		},
		{
			name: "unknown API",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"))),
			args: common.WebsocketSendArgs{Api: "bar", ChannelName: "/chat"},
			err:  "websocket API 'bar' not found",
		},
		{
			name: "unknown channel with API",
			app:  runtimetest.NewApp(runtimetest.WithWebsocket(newConfig("foo", "/chat"))),
			args: common.WebsocketSendArgs{Api: "foo", ChannelName: "/x"},
			err:  "websocket channel '/x' not found",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.args
			args.Message = message

			result, err := engine.NewWebsocketClient(tc.app).Send(args)

			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "foo", result.Api)
			expected := tc.channel
			if expected == "" {
				expected = "/chat"
			}
			require.Equal(t, expected, result.Channel)
		})
	}
}
