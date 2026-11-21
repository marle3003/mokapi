package websocket_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	engine "mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	ws "github.com/coder/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	chatCfg := func() *asyncapi3.Config {
		msg := asyncapi3test.NewMessage(
			asyncapi3test.WithContentType("application/json"),
			asyncapi3test.WithPayload(
				schematest.New("object",
					schematest.WithProperty("text", schematest.New("string")),
					schematest.WithRequired("text"),
				),
			),
		)
		ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}))

		return asyncapi3test.NewConfig(
			asyncapi3test.AddChannel("/chat", ch),
			asyncapi3test.WithOperation("sendChatMessage",
				asyncapi3test.WithOperationAction("send"),
				asyncapi3test.WithOperationChannel(ch),
				asyncapi3test.UseOperationMessage(msg),
			),
		)
	}

	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		js   string
		test func(t *testing.T, store *websocket.Store, url string, app *runtime.App)
	}{
		{
			name: "channel exists",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("/subscribe"),
			),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/subscribe", nil)
				require.NoError(t, err)
				require.NotNil(t, c)
				defer func() { _ = c.CloseNow() }()
			},
		},
		{
			name: "channel next to empty channel",
			cfg: func() *asyncapi3.Config {
				cfg := asyncapi3test.NewConfig(asyncapi3test.WithChannel("/subscribe"))
				cfg.Channels["/empty"] = nil
				require.NoError(t, cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{}))
				return cfg
			}(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/subscribe", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				_, _, err = ws.Dial(ctx, url+"/empty", nil)
				require.ErrorContains(t, err, "404")
			},
		},
		{
			name: "channel does not exist",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("/subscribe"),
			),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				_, _, err := ws.Dial(ctx, url+"/foo", nil)
				require.ErrorContains(t, err, "404")
			},
		},
		{
			name: "send valid message",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("text", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = c.Write(ctx, ws.MessageText, []byte(`{"text":"foo"}`))
				require.NoError(t, err)

				// Try to read with a short deadline — we expect it to time out, not close
				readCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()

				_, _, err = c.Read(readCtx)

				// A timeout means the connection is still alive — that's the success case
				// A close frame would mean the server rejected the message
				if ws.CloseStatus(err) != -1 {
					t.Fatalf("server closed connection unexpectedly: %v", err)
				}

				// err will be context.DeadlineExceeded — that's expected and means success
				assert.ErrorIs(t, err, context.DeadlineExceeded)
			},
		},
		{
			name: "send invalid message",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("integer"),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = c.Write(ctx, ws.MessageText, []byte("hello"))
				require.NoError(t, err)

				_, _, err = c.Read(ctx)
				require.ErrorContains(t, err, "invalid json format: invalid character 'h' looking for beginning of value")

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 2
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 3)

				evt := evts[0]
				cl := evt.Data.(*websocket.CloseLog)
				require.Equal(t, "connection", evt.Traits.Get("type"))
				require.Equal(t, "invalid json format: invalid character 'h' looking for beginning of value", cl.Reason)
				require.Equal(t, `server`, cl.ClosedBy)

				evt = evts[1]
				m := evt.Data.(*websocket.MessageLog)
				require.Equal(t, "message", evt.Traits.Get("type"))
				require.Equal(t, map[string]string{"ChatMessage": "invalid json format: invalid character 'h' looking for beginning of value"}, m.ValidationErrors)
			},
		},
		{
			name: "events are triggered",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		   	console.log(event.type)
			if (event.type === 'close') {
				console.log(event.reason, event.closedBy)
			}
		  }, { track: true })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				log.Info("url2", url)

				// Send a valid message
				err = c.Write(ctx, ws.MessageText, []byte(`{"text": "hello2"}`))
				require.NoError(t, err)

				_ = c.Close(ws.StatusNormalClosure, "bye")

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 2 && handlersFinished(evts)
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 3)

				evt := evts[0]
				cl := evt.Data.(*websocket.CloseLog)
				require.Equal(t, "connection", evt.Traits.Get("type"))
				require.Len(t, cl.Actions, 1)
				require.Nil(t, cl.Actions[0].Error)
				require.Equal(t, `close`, cl.Actions[0].Logs[0].Message)
				require.Equal(t, `bye client`, cl.Actions[0].Logs[1].Message)

				evt = evts[1]
				m := evt.Data.(*websocket.MessageLog)
				require.Equal(t, "message", evt.Traits.Get("type"))
				require.Len(t, m.GetActions(), 1)
				require.Equal(t, `message`, m.GetActions()[0].Logs[0].Message)

				evt = evts[2]
				d := evt.Data.(*websocket.ConnectLog)
				require.Equal(t, "connection", evt.Traits.Get("type"))
				require.Len(t, d.GetActions(), 1)
				require.Equal(t, `connect`, d.GetActions()[0].Logs[0].Message)
			},
		},
		{
			name: "reply",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
export default function() {
  on('websocket', function(event) {
    event.reply({text: "pong"})
  })
}
`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.Close(ws.StatusNormalClosure, "") }()

				log.Info("url1", url)

				// Send a valid message
				err = c.Write(ctx, ws.MessageText, []byte(`{"text": "ping"}`))
				require.NoError(t, err)

				mt, data, err := c.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, mt, ws.MessageText)
				require.Equal(t, `{"text":"pong"}`, string(data))
			},
		},
		{
			name: "reply not valid",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
			if (event.type === 'message') {
		    	event.reply({text2: "hello"})
			}
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				// Send a valid message
				err = c.Write(ctx, ws.MessageText, []byte(`{"text": "hello"}`))
				require.NoError(t, err)

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 2 && handlersFinished(evts)
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 3)
				evt := evts[0]
				d := evt.Data.(*websocket.MessageLog)
				require.Len(t, d.GetActions(), 0)
				require.Equal(t, map[string]string{"ChatMessage": "Validation error count 1:\n\t- #/required: required properties are missing: text"}, d.ValidationErrors)

				evt = evts[1]
				d = evt.Data.(*websocket.MessageLog)
				require.Len(t, d.GetActions(), 1)
				require.NotNil(t, d.GetActions()[0].Error)
				require.Equal(t, "Validation error count 1:\n\t- #/required: required properties are missing: text", d.GetActions()[0].Error.Message)
			},
		},
		{
			name: "use client.send",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		    event.client.send({text: "hello"})
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				// Send a valid message
				err = c.Write(ctx, ws.MessageText, []byte(`{"text": "hello"}`))
				require.NoError(t, err)

				mt, data, err := c.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, mt, ws.MessageText)
				require.Equal(t, `{"text":"hello"}`, string(data))
			},
		},
		{
			name: "use broadcast",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		    event.broadcast({text: "broadcast"})
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c1, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c1.CloseNow() }()

				c2, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c2.CloseNow() }()

				// Send a valid message
				err = c1.Write(ctx, ws.MessageText, []byte(`{"text": "ping"}`))
				require.NoError(t, err)

				mt, data, err := c1.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, mt, ws.MessageText)
				require.Equal(t, `{"text":"broadcast"}`, string(data))

				mt, data, err = c2.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, mt, ws.MessageText)
				require.Equal(t, `{"text":"broadcast"}`, string(data))
			},
		},
		{
			name: "binding query id",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithWebsocketChannelBinding(asyncapi3.WebsocketChannelBindings{
						Query: schematest.New("object",
							schematest.WithProperty("id", schematest.New("string")),
						),
					}),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		    console.log(event.client.query)
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat?id=foo", &ws.DialOptions{})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = c.Write(ctx, ws.MessageText, []byte("hello"))
				require.NoError(t, err)

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 1 && handlersFinished(evts)
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 2)
				evt := evts[0]
				d := evt.Data.(*websocket.MessageLog)
				require.Len(t, d.GetActions(), 1)
				require.Equal(t, `{"id":"foo"}`, d.GetActions()[0].Logs[0].Message)
			},
		},
		{
			name: "binding query id required",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithWebsocketChannelBinding(asyncapi3.WebsocketChannelBindings{
						Query: schematest.New("object",
							schematest.WithProperty("id", schematest.New("string")),
							schematest.WithRequired("id"),
						),
					}),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				_, _, err := ws.Dial(ctx, url+"/chat", &ws.DialOptions{})
				require.ErrorContains(t, err, "400")
			},
		},
		{
			name: "binding header id",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithWebsocketChannelBinding(asyncapi3.WebsocketChannelBindings{
						Headers: schematest.New("object",
							schematest.WithProperty("id", schematest.New("string")),
						),
					}),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
			console.log(event.client.headers)
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", &ws.DialOptions{
					HTTPHeader: http.Header{
						"id": []string{"foo"},
					},
				})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = c.Write(ctx, ws.MessageText, []byte("hello"))
				require.NoError(t, err)

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 1 && handlersFinished(evts)
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 2)

				evt := evts[0]
				m := evt.Data.(*websocket.MessageLog)
				require.Equal(t, "message", evt.Traits.Get("type"))
				require.Len(t, m.GetActions(), 1)
				require.Equal(t, `{"id":"foo"}`, m.GetActions()[0].Logs[0].Message)
			},
		},
		{
			name: "binding header id required",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithWebsocketChannelBinding(asyncapi3.WebsocketChannelBindings{
						Headers: schematest.New("object",
							schematest.WithProperty("id", schematest.New("string")),
							schematest.WithRequired("id"),
						),
					}),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chat", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				_, _, err := ws.Dial(ctx, url+"/chat", &ws.DialOptions{})
				require.ErrorContains(t, err, "400")
			},
		},
		{
			name: "channel with parameter",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("ChatMessage", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithParameter(
						"chatId", &asyncapi3.Parameter{},
					),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("/chats/{chatId}", ch),
					asyncapi3test.WithOperation("sendChatMessage",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		    console.log(event.channel)
		  })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chats/1234", &ws.DialOptions{})
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				err = c.Write(ctx, ws.MessageText, []byte("hello"))
				require.NoError(t, err)

				try.WaitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 1 && handlersFinished(evts)
				})

				require.Equal(t, 1.0, app.Monitor.Websocket.Messages.Sum(
					metrics.NewQuery(metrics.ByLabel("channel", "/chats/{chatId}")),
				))
				require.Equal(t, 1.0, app.Monitor.Websocket.Messages.Sum(
					metrics.NewQuery(
						metrics.ByLabel("channel", "/chats/{chatId}"),
						metrics.ByLabel("instance", "/chats/1234"),
					),
				))

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Equal(t, "/chats/{chatId}", evts[0].Traits.Get("channel"))
				require.Equal(t, "/chats/1234", evts[0].Traits.Get("instance"))
				require.Len(t, evts, 2)
				evt := evts[0]
				d := evt.Data.(*websocket.MessageLog)
				require.Len(t, d.GetActions(), 1)
				var result engine.WebsocketEventChannel
				err = json.Unmarshal([]byte(d.GetActions()[0].Logs[0].Message), &result)
				require.NoError(t, err)
				require.Equal(t, "/chats/1234", result.Name)
				require.Len(t, result.Clients, 1)
				require.NotNil(t, result.Clients[0])
				require.Equal(t, map[string]string{"chatId": "1234"}, result.Params)
			},
		},
		{
			name: "broadcast right after handshake reaches client",
			cfg:  chatCfg(),
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()

				// no waiting: the client must already be known once Dial returns
				ch, ok := store.Channel("/chat")
				require.True(t, ok)
				require.NoError(t, ch.Broadcast(map[string]any{"text": "hello"}))

				_, data, err := c.Read(ctx)
				require.NoError(t, err)
				require.JSONEq(t, `{"text":"hello"}`, string(data))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			log.SetOutput(io.Discard)

			app := runtimetest.NewApp()
			e := enginetest.NewEngine()
			app.Engine = e

			defer e.Close()

			if tc.js != "" {
				err := e.AddScript(dynamic.ConfigEvent{
					Config: &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: try.MustUrl("./script.ts")},
						Raw:  []byte(tc.js),
						Data: tc.js,
					},
					Event: dynamic.Create,
				})
				require.NoError(t, err)
			}

			s := websocket.New(tc.cfg, app.Engine, app.Events, app.Monitor.Websocket)

			serv := httptest.NewServer(s)
			defer serv.Close()

			tc.test(t, s, serv.URL, app)
		})
	}
}

// handlersFinished reports whether the script handlers of all message events have attached
// their actions. The event is visible in the store before its handler has finished.
func handlersFinished(evts []events.Event) bool {
	for _, evt := range evts {
		switch l := evt.Data.(type) {
		case *websocket.MessageLog:
			if len(l.ValidationErrors) == 0 && len(l.GetActions()) == 0 {
				return false
			}
		}
	}
	return true
}

// Concurrent first requests for the same parameterized path must share one channel,
// otherwise a broadcast only reaches the clients of the sender's channel.
func TestStore_Channel_ConcurrentFirstRequestsShareChannel(t *testing.T) {
	const iterations = 100
	const clients = 8

	for i := 0; i < iterations; i++ {
		ch := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
		)
		cfg := asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", ch))
		s := websocket.New(cfg, nil, nil, nil)

		start := make(chan struct{})
		got := make([]*websocket.Channel, clients)
		var wg sync.WaitGroup
		for j := 0; j < clients; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				got[j], _ = s.Channel("/chats/general")
			}()
		}
		close(start)
		wg.Wait()

		for j := 1; j < clients; j++ {
			require.Same(t, got[0], got[j], "iteration %d: client %d got a different channel", i, j)
		}
	}
}

func TestStore_Clients_ConcurrentWithFirstChannelRequests(t *testing.T) {
	const iterations = 100
	const requests = 8

	for i := 0; i < iterations; i++ {
		ch := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
		)
		cfg := asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", ch))
		s := websocket.New(cfg, nil, nil, nil)

		start := make(chan struct{})
		var wg sync.WaitGroup
		for j := 0; j < requests; j++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, _ = s.Channel(fmt.Sprintf("/chats/%d", j))
			}()
			go func() {
				defer wg.Done()
				<-start
				_ = s.Clients()
			}()
		}
		close(start)
		wg.Wait()
	}
}

func TestStore_Update_ConcurrentWithChannel(t *testing.T) {
	const iterations = 100
	const requests = 8

	newCfg := func() *asyncapi3.Config {
		ch := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
		)
		return asyncapi3test.NewConfig(
			asyncapi3test.AddChannel("/chats/{roomId}", ch),
			asyncapi3test.WithChannel("/static"),
		)
	}

	for i := 0; i < iterations; i++ {
		s := websocket.New(newCfg(), nil, nil, nil)
		reloaded := newCfg()

		start := make(chan struct{})
		var wg sync.WaitGroup
		for j := 0; j < requests; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, _ = s.Channel(fmt.Sprintf("/chats/%d", j))
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			s.Update(reloaded)
		}()
		close(start)
		wg.Wait()
	}
}

func TestStore_Update_ConcurrentWithConfigRead(t *testing.T) {
	const iterations = 50

	newCfg := func() *asyncapi3.Config {
		room := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("room"),
		)
		chat := asyncapi3test.NewChannel(asyncapi3test.WithMessage("chat"))
		return asyncapi3test.NewConfig(
			asyncapi3test.AddChannel("/rooms/{roomId}", room),
			asyncapi3test.AddChannel("/chat", chat),
		)
	}

	for i := 0; i < iterations; i++ {
		s := websocket.New(newCfg(), nil, nil, nil)
		ch, ok := s.Channel("/chat")
		require.True(t, ok)

		stop := make(chan struct{})
		var wg sync.WaitGroup
		for _, path := range []string{"/chat", "/rooms/1"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
					}
					if c, ok := s.Channel(path); ok {
						_ = c.Config().ResolveAddress()
					}
					rr := httptest.NewRecorder()
					s.ServeHTTP(rr, httptest.NewRequest("GET", path+"?a=b", nil))
				}
			}()
		}

		for j := 0; j < 20; j++ {
			s.Update(newCfg())
			_ = ch.Config()
		}
		close(stop)
		wg.Wait()
	}
}

func TestStore_Update_Reload(t *testing.T) {
	t.Run("changed channel uses new config", func(t *testing.T) {
		oldCh := asyncapi3test.NewChannel(asyncapi3test.WithMessage("old"))
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chat", oldCh)), nil, nil, nil)

		newCh := asyncapi3test.NewChannel(asyncapi3test.WithMessage("new"))
		s.Update(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chat", newCh)))

		ch, ok := s.Channel("/chat")
		require.True(t, ok)
		require.Same(t, newCh, ch.Config())
	})

	t.Run("removed channel is not served", func(t *testing.T) {
		s := websocket.New(asyncapi3test.NewConfig(
			asyncapi3test.WithChannel("/chat"),
			asyncapi3test.WithChannel("/other"),
		), nil, nil, nil)

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		_, ok := s.Channel("/chat")
		require.False(t, ok)
		rr := httptest.NewRecorder()
		s.ServeHTTP(rr, httptest.NewRequest("GET", "/chat", nil))
		require.Equal(t, http.StatusNotFound, rr.Code)
		require.Len(t, s.Channels(), 1)
	})

	t.Run("parameterized channel uses new config", func(t *testing.T) {
		oldCh := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("old"))
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", oldCh)), nil, nil, nil)
		_, ok := s.Channel("/chats/general")
		require.True(t, ok)

		newCh := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("new"))
		s.Update(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", newCh)))

		ch, ok := s.Channel("/chats/general")
		require.True(t, ok)
		require.Same(t, newCh, ch.Config())
	})

	t.Run("removed parameterized channel is not served", func(t *testing.T) {
		ch := asyncapi3test.NewChannel(asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}))
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", ch)), nil, nil, nil)
		_, ok := s.Channel("/chats/general")
		require.True(t, ok)

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		_, ok = s.Channel("/chats/general")
		require.False(t, ok)
	})

	t.Run("channel with address different from its key uses new config", func(t *testing.T) {
		oldCh := asyncapi3test.NewChannel(asyncapi3test.WithChannelAddress("/chat"), asyncapi3test.WithMessage("old"))
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("chat", oldCh)), nil, nil, nil)

		newCh := asyncapi3test.NewChannel(asyncapi3test.WithChannelAddress("/chat"), asyncapi3test.WithMessage("new"))
		s.Update(asyncapi3test.NewConfig(asyncapi3test.AddChannel("chat", newCh)))

		ch, ok := s.Channel("/chat")
		require.True(t, ok)
		require.Same(t, newCh, ch.Config())
	})

	t.Run("channel keeps its identity", func(t *testing.T) {
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")), nil, nil, nil)
		before, _ := s.Channel("/chat")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")))

		after, ok := s.Channel("/chat")
		require.True(t, ok)
		require.Same(t, before, after)
	})

	t.Run("parameterized channel gets the new parameter names", func(t *testing.T) {
		oldCh := asyncapi3test.NewChannel(asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}))
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chats/{roomId}", oldCh)), nil, nil, nil)
		_, ok := s.Channel("/chats/general")
		require.True(t, ok)

		newCh := asyncapi3test.NewChannel(
			asyncapi3test.WithChannelAddress("/chats/{room}"),
			asyncapi3test.WithParameter("room", &asyncapi3.Parameter{}))
		s.Update(asyncapi3test.NewConfig(asyncapi3test.AddChannel("chats", newCh)))

		ch, ok := s.Channel("/chats/general")
		require.True(t, ok)
		require.Same(t, newCh, ch.Config())
		require.Equal(t, map[string]string{"room": "general"}, websocket.ChannelParams(ch))
	})

	t.Run("removed and re-added channel is a new channel without the old clients", func(t *testing.T) {
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")), nil, nil, nil)
		before, _ := s.Channel("/chat")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))
		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")))

		after, ok := s.Channel("/chat")
		require.True(t, ok)
		require.NotSame(t, before, after)
	})

	t.Run("channel no longer available for ws is removed", func(t *testing.T) {
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")), nil, nil, nil)

		kafkaOnly := asyncapi3test.NewChannel()
		kafkaOnly.Servers = []*asyncapi3.ServerRef{{Value: &asyncapi3.Server{Protocol: "kafka"}}}
		s.Update(asyncapi3test.NewConfig(asyncapi3test.AddChannel("/chat", kafkaOnly)))

		_, ok := s.Channel("/chat")
		require.False(t, ok)
	})

	t.Run("nil channel reference removes the channel", func(t *testing.T) {
		s := websocket.New(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/chat")), nil, nil, nil)

		cfg := asyncapi3test.NewConfig()
		cfg.Channels = map[string]*asyncapi3.ChannelRef{"/chat": nil}
		s.Update(cfg)

		_, ok := s.Channel("/chat")
		require.False(t, ok)
	})
}

func TestStore_Update_ClosesClientsOfRemovedChannels(t *testing.T) {
	log.SetOutput(io.Discard)

	start := func(t *testing.T, cfg *asyncapi3.Config, js string) (*websocket.Store, string, *runtime.App) {
		app := runtimetest.NewApp()
		e := enginetest.NewEngine()
		app.Engine = e
		t.Cleanup(e.Close)

		if js != "" {
			err := e.AddScript(dynamic.ConfigEvent{
				Config: &dynamic.Config{
					Info: dynamic.ConfigInfo{Url: try.MustUrl("./script.ts")},
					Raw:  []byte(js),
					Data: js,
				},
				Event: dynamic.Create,
			})
			require.NoError(t, err)
		}

		s := websocket.New(cfg, app.Engine, app.Events, app.Monitor.Websocket)
		serv := httptest.NewServer(s)
		t.Cleanup(serv.Close)
		return s, serv.URL, app
	}

	dial := func(t *testing.T, url string) *ws.Conn {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c, _, err := ws.Dial(ctx, url, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = c.CloseNow() })
		return c
	}

	twoChannels := func() *asyncapi3.Config {
		return asyncapi3test.NewConfig(
			asyncapi3test.WithChannel("/chat"),
			asyncapi3test.WithChannel("/other"),
		)
	}

	t.Run("client of removed channel gets close status going away", func(t *testing.T) {
		s, url, _ := start(t, twoChannels(), "")
		c := dial(t, url+"/chat")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _, err := c.Read(ctx)
		closeErr, ok := errors.AsType[ws.CloseError](err)
		require.True(t, ok, "expected close error, got %v", err)
		require.Equal(t, ws.StatusGoingAway, closeErr.Code)
		require.Equal(t, "channel removed", closeErr.Reason)
	})

	t.Run("client of parameterized channel instance is closed", func(t *testing.T) {
		ch := asyncapi3test.NewChannel(asyncapi3test.WithParameter("roomId", &asyncapi3.Parameter{}))
		s, url, _ := start(t, asyncapi3test.NewConfig(asyncapi3test.AddChannel("/rooms/{roomId}", ch)), "")
		c := dial(t, url+"/rooms/1")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _, err := c.Read(ctx)
		require.Equal(t, ws.StatusGoingAway, ws.CloseStatus(err))
	})

	t.Run("close is logged once as closed by server and runs close event", func(t *testing.T) {
		js := `import { on } from 'mokapi'
export default function() {
  on('websocket', function(event) {
    if (event.type === 'close') {
      console.log(event.reason, event.closedBy)
    }
  })
}
`
		s, url, app := start(t, twoChannels(), js)
		c := dial(t, url+"/chat")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _, _ = c.Read(ctx)

		closeLogs := func() []*websocket.CloseLog {
			var logs []*websocket.CloseLog
			for _, evt := range app.Events.GetEvents(events.NewTraits().WithNamespace("websocket")) {
				if cl, ok := evt.Data.(*websocket.CloseLog); ok {
					logs = append(logs, cl)
				}
			}
			return logs
		}
		try.WaitFor(t, func() bool { return len(closeLogs()) > 0 })
		time.Sleep(100 * time.Millisecond)

		logs := closeLogs()
		require.Len(t, logs, 1)
		require.Equal(t, "server", logs[0].ClosedBy)
		require.Equal(t, "channel removed", logs[0].Reason)
		require.Len(t, logs[0].Actions, 1)
		require.Equal(t, "channel removed server", logs[0].Actions[0].Logs[0].Message)
	})

	t.Run("client of kept channel stays connected", func(t *testing.T) {
		s, url, _ := start(t, twoChannels(), "")
		c := dial(t, url+"/other")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		// a read timeout closes the connection, so check the client first
		require.Len(t, s.Clients(), 1)
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		_, _, err := c.Read(ctx)
		require.ErrorIs(t, err, context.DeadlineExceeded, "connection must stay open")
	})

	t.Run("closing by client and by server at the same time", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			s, url, _ := start(t, twoChannels(), "")
			c := dial(t, url+"/chat")

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = c.Close(ws.StatusNormalClosure, "bye")
			}()
			s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))
			<-done

			try.WaitFor(t, func() bool { return len(s.Clients()) == 0 })
		}
	})

	t.Run("client added after the channel was removed is closed", func(t *testing.T) {
		s, _, _ := start(t, twoChannels(), "")
		ch, ok := s.Channel("/chat")
		require.True(t, ok)

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		reason, closedByServer := websocket.NewClientOnChannel(ch)
		require.True(t, closedByServer)
		require.Equal(t, "channel removed", reason)
	})

	t.Run("client added to a kept channel is not closed", func(t *testing.T) {
		s, _, _ := start(t, twoChannels(), "")
		ch, _ := s.Channel("/other")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		_, closedByServer := websocket.NewClientOnChannel(ch)
		require.False(t, closedByServer)
	})

	t.Run("update without clients", func(t *testing.T) {
		s, _, _ := start(t, twoChannels(), "")

		s.Update(asyncapi3test.NewConfig(asyncapi3test.WithChannel("/other")))

		require.Len(t, s.Channels(), 1)
	})
}
