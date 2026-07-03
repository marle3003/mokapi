package websocket_test

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/websocket"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
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
					asyncapi3test.WithContentType("text/plain"),
					asyncapi3test.WithPayload(
						schematest.New("string"),
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
			},
		},
		{
			name: "event is triggered",
			cfg:  chatCfg(),
			js: `import { on } from 'mokapi'
		export default function() {
		  on('websocket', function(event) {
		    console.log(event.message)
		  }, { track: true })
		}
		`,
			test: func(t *testing.T, store *websocket.Store, url string, app *runtime.App) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				c, _, err := ws.Dial(ctx, url+"/chat", nil)
				require.NoError(t, err)
				defer func() { _ = c.CloseNow() }()
				log.Info("url2", url)

				// Send a valid message
				err = c.Write(ctx, ws.MessageText, []byte(`{"text": "hello2"}`))
				require.NoError(t, err)

				waitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 0
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*websocket.Log)
				require.Len(t, d.Actions, 1)
				require.Equal(t, `{"text":"hello2"}`, d.Actions[0].Logs[0].Message)
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
		    event.reply({text2: "hello"})
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

				waitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 0
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*websocket.Log)
				require.Len(t, d.Actions, 1)
				require.NotNil(t, d.Actions[0].Error)
				require.Equal(t, "Validation error count 1:\n\t- #/required: required properties are missing: text", d.Actions[0].Error.Message)
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

				waitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 0
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*websocket.Log)
				require.Len(t, d.Actions, 1)
				require.Equal(t, `{"id":"foo"}`, d.Actions[0].Logs[0].Message)
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

				waitFor(t, func() bool {
					evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
					return len(evts) > 0
				})

				evts := app.Events.GetEvents(events.NewTraits().WithNamespace("websocket"))
				require.Len(t, evts, 1)
				evt := evts[0]
				d := evt.Data.(*websocket.Log)
				require.Len(t, d.Actions, 1)
				require.Equal(t, `{"id":"foo"}`, d.Actions[0].Logs[0].Message)
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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

func waitFor(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)

	for {
		if check() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("waitFor deadline")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
