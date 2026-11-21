package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"mokapi/api"
	"mokapi/config/static"
	"mokapi/try"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

type WebsocketSuite struct{ BaseSuite }

func (suite *WebsocketSuite) SetupSuite() {
	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: "./websocket"}}
	suite.initCmd(cfg)
}

// TestConnect verifies that a single client can connect to a chat room.
func (suite *WebsocketSuite) TestConnect() {
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	err = c.Close(websocket.StatusNormalClosure, "")
	require.NoError(suite.T(), err)
}

// TestEcho verifies that the sender receives back a message sent to a chat room.
func (suite *WebsocketSuite) TestEcho() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	err = wsjson.Write(ctx, c, map[string]string{
		"user": "alice",
		"text": "Hello, world!",
	})
	require.NoError(suite.T(), err)

	var msg map[string]any
	err = wsjson.Read(ctx, c, &msg)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "alice", msg["user"])
	require.Equal(suite.T(), "Hello, world!", msg["text"])
	require.NotEmpty(suite.T(), msg["timestamp"])
}

// TestSendAsyncSendsSnapshotOfMessage verifies that sendAsync sends the message as it was at
// call time, even if the script changes the message object right after the call.
func (suite *WebsocketSuite) TestSendAsyncSendsSnapshotOfMessage() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/async", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	err = wsjson.Write(ctx, c, map[string]string{"text": "trigger"})
	require.NoError(suite.T(), err)

	var msg map[string]any
	err = wsjson.Read(ctx, c, &msg)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "hello", msg["text"])
}

// TestBroadcast verifies that a message is delivered to all connected clients in the same room.
func (suite *WebsocketSuite) TestBroadcast() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect two clients to the same room
	c1, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer c1.CloseNow()

	c2, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer c2.CloseNow()

	// Send message from c1
	err = wsjson.Write(ctx, c1, map[string]string{
		"user": "alice",
		"text": "Hello everyone!",
	})
	require.NoError(suite.T(), err)

	// Both c1 and c2 should receive the broadcast
	var wg sync.WaitGroup
	wg.Add(2)
	results := make([]map[string]any, 2)
	errs := make([]error, 2)

	for i, c := range []*websocket.Conn{c1, c2} {
		i, c := i, c
		go func() {
			defer wg.Done()
			errs[i] = wsjson.Read(ctx, c, &results[i])
		}()
	}
	wg.Wait()

	for i := range results {
		require.NoError(suite.T(), errs[i], "client %d read error", i)
		require.Equal(suite.T(), "alice", results[i]["user"], "client %d user", i)
		require.Equal(suite.T(), "Hello everyone!", results[i]["text"], "client %d text", i)
	}
}

// TestRoomIsolation verifies that messages in one room are not delivered to a different room.
func (suite *WebsocketSuite) TestRoomIsolation() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect one client to "general", another to "random"
	cGeneral, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer cGeneral.CloseNow()

	cRandom, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/random", nil)
	require.NoError(suite.T(), err)
	defer cRandom.CloseNow()

	// Send to "general" room
	err = wsjson.Write(ctx, cGeneral, map[string]string{
		"user": "alice",
		"text": "Only for general",
	})
	require.NoError(suite.T(), err)

	// cGeneral should receive the message
	var msg map[string]any
	err = wsjson.Read(ctx, cGeneral, &msg)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Only for general", msg["text"])

	// cRandom should NOT receive any message within a short window
	cRandom.SetReadLimit(512)
	readCtx, readCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer readCancel()
	_, _, err = cRandom.Read(readCtx)
	require.ErrorIs(suite.T(), err, context.DeadlineExceeded,
		"expected no message in the 'random' room, but got one")
}

// TestDirectMessage verifies that a direct message receives an acknowledgement reply.
func (suite *WebsocketSuite) TestDirectMessage() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/dm", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	err = wsjson.Write(ctx, c, map[string]string{
		"to":   "bob",
		"text": "Hey Bob, are you there?",
	})
	require.NoError(suite.T(), err)

	var ack map[string]any
	err = wsjson.Read(ctx, c, &ack)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "ok", ack["status"])
}

// TestMultipleRoomsParallel verifies that multiple rooms work independently and concurrently.
func (suite *WebsocketSuite) TestMultipleRoomsParallel() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rooms := []string{"room-a", "room-b", "room-c"}
	var wg sync.WaitGroup

	for _, room := range rooms {
		room := room
		wg.Add(1)
		go func() {
			defer wg.Done()

			c, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/"+room, nil)
			require.NoError(suite.T(), err)
			defer c.CloseNow()

			err = wsjson.Write(ctx, c, map[string]string{
				"user": "bot-" + room,
				"text": "message in " + room,
			})
			require.NoError(suite.T(), err)

			var msg map[string]any
			err = wsjson.Read(ctx, c, &msg)
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), "message in "+room, msg["text"])
		}()
	}

	wg.Wait()
}

// TestMessageOrder verifies that multiple messages are delivered in order.
func (suite *WebsocketSuite) TestMessageOrder() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/order-test", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	messages := []string{"first", "second", "third"}
	for _, text := range messages {
		err = wsjson.Write(ctx, c, map[string]string{
			"user": "alice",
			"text": text,
		})
		require.NoError(suite.T(), err)
	}

	for _, expected := range messages {
		var msg map[string]any
		err = wsjson.Read(ctx, c, &msg)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), expected, msg["text"])
	}
}

// TestPayloadValidation verifies that an invalid message payload is rejected.
func (suite *WebsocketSuite) TestPayloadValidation() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/chat/general", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	// Send payload missing required "user" field
	invalidPayload, _ := json.Marshal(map[string]string{
		"text": "missing user field",
	})
	err = c.Write(ctx, websocket.MessageText, invalidPayload)
	require.NoError(suite.T(), err)

	// Expect the connection to be closed with a policy violation or an error message
	_, _, err = c.Read(ctx)
	require.Error(suite.T(), err, "expected error or close after invalid payload")
}

// postSimulation asks the fixture script to send a message to the channel /dm.
func (suite *WebsocketSuite) postSimulation(endpoint, body string, conditions ...try.ResponseCondition) {
	waitForScript(suite.T(), "http://localhost:18096")
	try.Request(suite.T(), http.MethodPost, "http://localhost:18096"+endpoint,
		map[string]string{"Content-Type": "application/json"}, strings.NewReader(body), conditions...)
}

// simulateSend connects a client to /dm, asks the fixture script to send a message to the channel
// and returns what the client receives. The conditions check the script's HTTP response.
func (suite *WebsocketSuite) simulateSend(endpoint, body string, conditions ...try.ResponseCondition) map[string]any {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/dm", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	suite.postSimulation(endpoint, body, conditions...)
	if suite.T().Failed() {
		return nil
	}

	var msg map[string]any
	err = wsjson.Read(ctx, c, &msg)
	require.NoError(suite.T(), err)
	return msg
}

// TestSendGeneratesSelectedMessage verifies that messageId without a message sends a payload
// generated from the selected message of the channel.
func (suite *WebsocketSuite) TestSendGeneratesSelectedMessage() {
	msg := suite.simulateSend("/sim/websocket", `{"messageId":"directMessage"}`, try.HasStatusCode(http.StatusNoContent))
	require.Contains(suite.T(), msg, "to")
	require.Contains(suite.T(), msg, "text")

	msg = suite.simulateSend("/sim/websocket", `{"messageId":"ack"}`, try.HasStatusCode(http.StatusNoContent))
	require.Contains(suite.T(), []any{"ok", "error"}, msg["status"])
}

// TestSendAsyncGeneratesSelectedMessage verifies that sendAsync selects the message by messageId.
func (suite *WebsocketSuite) TestSendAsyncGeneratesSelectedMessage() {
	msg := suite.simulateSend("/sim/websocket-async", `{"messageId":"directMessage"}`, try.HasStatusCode(http.StatusNoContent))

	require.Contains(suite.T(), msg, "to")
	require.Contains(suite.T(), msg, "text")
}

// TestSendMessageIsValidatedAgainstSelectedMessage verifies that a message valid for the selected
// message is sent as given.
func (suite *WebsocketSuite) TestSendMessageIsValidatedAgainstSelectedMessage() {
	msg := suite.simulateSend("/sim/websocket", `{"messageId":"ack","message":{"status":"error","error":"boom"}}`,
		try.HasStatusCode(http.StatusNoContent))

	require.Equal(suite.T(), map[string]any{"status": "error", "error": "boom"}, msg)
}

// TestSendInvalidMessageForSelectedMessageFails verifies that a message not matching the selected
// message is rejected.
func (suite *WebsocketSuite) TestSendInvalidMessageForSelectedMessageFails() {
	suite.postSimulation("/sim/websocket", `{"messageId":"directMessage","message":{"status":"ok"}}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("required properties are missing"))
}

// TestSendUnknownMessageIdFails verifies that a typo in messageId is reported to the script.
func (suite *WebsocketSuite) TestSendUnknownMessageIdFails() {
	suite.postSimulation("/sim/websocket", `{"messageId":"ackk"}`,
		try.HasStatusCode(http.StatusUnprocessableEntity),
		try.BodyContains("message 'ackk' not found in channel '/dm'"))
}

// messageEvents returns the data of all message events of the websocket namespace.
func (suite *WebsocketSuite) messageEvents() []map[string]any {
	var page api.EventPage
	suite.getJSON("/api/events?traits[namespace]=websocket&traits[type]=message&size=1000", &page)

	var all []map[string]any
	for _, e := range page.Items {
		all = append(all, e.Data.(map[string]any))
	}
	return all
}

// configId returns the id the dashboard uses to link a config file, by the end of its url.
func (suite *WebsocketSuite) configId(urlSuffix string) string {
	var configs []struct {
		Id  string `json:"id"`
		Url string `json:"url"`
	}
	suite.getJSON("/api/configs", &configs)
	for _, c := range configs {
		if strings.HasSuffix(c.Url, urlSuffix) {
			return c.Id
		}
	}
	suite.T().Fatalf("config %v not found", urlSuffix)
	return ""
}

func (suite *WebsocketSuite) getJSON(path string, v any) {
	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", suite.cfg.Api.Port, path))
	require.NoError(suite.T(), err)
	defer res.Body.Close()
	require.NoError(suite.T(), json.NewDecoder(res.Body).Decode(v))
}

func messageContaining(events []map[string]any, text string) []map[string]any {
	var found []map[string]any
	for _, e := range events {
		if strings.Contains(e["message"].(map[string]any)["value"].(string), text) {
			found = append(found, e)
		}
	}
	return found
}

// TestMessageSentByScriptLogsScriptFile verifies that the message event of a message sent with
// mokapi.websocket().send() names the script that sent it.
func (suite *WebsocketSuite) TestMessageSentByScriptLogsScriptFile() {
	time.Sleep(2 * time.Second)

	suite.simulateSend("/sim/websocket", `{"messageId":"ack","message":{"status":"error","error":"logged-by-script"}}`,
		try.HasStatusCode(http.StatusNoContent))

	events := messageContaining(suite.messageEvents(), "logged-by-script")

	require.Len(suite.T(), events, 1)
	require.Equal(suite.T(), "receive", events[0]["direction"])
	require.Equal(suite.T(), suite.configId("/sim.ts"), events[0]["script"])
}

// TestMessageSentByHandlerHasNoScriptFile verifies that a reply sent by an event handler
// is logged without a script.
func (suite *WebsocketSuite) TestMessageSentByHandlerHasNoScriptFile() {
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "http://localhost:22801/dm", nil)
	require.NoError(suite.T(), err)
	defer c.CloseNow()

	require.NoError(suite.T(), wsjson.Write(ctx, c, map[string]string{"to": "bob", "text": "reply-by-handler"}))
	var reply map[string]any
	require.NoError(suite.T(), wsjson.Read(ctx, c, &reply))

	events := suite.messageEvents()
	sent := messageContaining(events, "reply-by-handler")
	require.Len(suite.T(), sent, 1)
	clientId := sent[0]["client"].(map[string]any)["id"]

	var received []map[string]any
	for _, e := range events {
		if e["direction"] == "receive" && e["client"].(map[string]any)["id"] == clientId {
			received = append(received, e)
		}
	}

	require.Len(suite.T(), received, 1)
	require.NotContains(suite.T(), received[0], "script")
}

// WebsocketReloadSuite checks what a mock developer sees when the AsyncAPI file is edited while
// Mokapi runs.
type WebsocketReloadSuite struct {
	BaseSuite
	dir  string
	port int
}

const (
	reloadChatChannel = `
  chat:
    address: /chat
`
	reloadChatChannelWithRequiredToken = `
  chat:
    address: /chat
    bindings:
      ws:
        query:
          type: object
          properties:
            token:
              type: string
          required: [token]
`
	reloadRoomsChannel = `
  rooms:
    address: /rooms/{roomId}
    parameters:
      roomId: {}
`
)

func (suite *WebsocketReloadSuite) SetupSuite() {
	suite.port = try.GetFreePort()
	suite.dir = suite.T().TempDir()
	suite.writeSpec(reloadChatChannel + reloadRoomsChannel)

	cfg := static.NewConfig()
	cfg.Api.Port = try.GetFreePort()
	cfg.Providers.File.Directories = []static.FileConfig{{Path: suite.dir}}
	suite.initCmd(cfg)
}

func (suite *WebsocketReloadSuite) writeSpec(channels string) {
	spec := fmt.Sprintf(`asyncapi: 3.0.0
info:
  title: Reload Chat
  version: 1.0.0
servers:
  mokapi:
    host: 'localhost:%d'
    protocol: ws
channels:%s`, suite.port, channels)
	suite.Require().NoError(os.WriteFile(filepath.Join(suite.dir, "reload.yml"), []byte(spec), 0o600))
}

// dial returns the connection (nil if the handshake failed) and the status code of the handshake.
func (suite *WebsocketReloadSuite) dial(path string) (*websocket.Conn, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, fmt.Sprintf("http://localhost:%d%s", suite.port, path), nil)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode
		}
		return nil, 0
	}
	return c, resp.StatusCode
}

func (suite *WebsocketReloadSuite) waitForStatus(path string, status int) {
	suite.T().Helper()
	try.WaitFor(suite.T(), func() bool {
		c, code := suite.dial(path)
		if c != nil {
			_ = c.CloseNow()
		}
		return code == status
	})
}

// reset serves the initial spec again and waits until both channels accept connections.
func (suite *WebsocketReloadSuite) reset() {
	suite.writeSpec(reloadChatChannel + reloadRoomsChannel)
	suite.waitForStatus("/chat", http.StatusSwitchingProtocols)
	suite.waitForStatus("/rooms/1", http.StatusSwitchingProtocols)
}

// connect opens a connection that stays open until the end of the test.
func (suite *WebsocketReloadSuite) connect(path string) *websocket.Conn {
	c, code := suite.dial(path)
	suite.Require().Equal(http.StatusSwitchingProtocols, code)
	suite.T().Cleanup(func() { _ = c.CloseNow() })
	return c
}

// TestChangedQueryBindingAppliesToNewConnections verifies that a query parameter made required
// in the file is enforced for the next connection.
func (suite *WebsocketReloadSuite) TestChangedQueryBindingAppliesToNewConnections() {
	suite.reset()

	suite.writeSpec(reloadChatChannelWithRequiredToken + reloadRoomsChannel)

	suite.waitForStatus("/chat", http.StatusBadRequest)
}

// TestRemovedChannelReturnsNotFound verifies that a channel removed from the file is not served.
func (suite *WebsocketReloadSuite) TestRemovedChannelReturnsNotFound() {
	suite.reset()

	suite.writeSpec(reloadRoomsChannel)

	suite.waitForStatus("/chat", http.StatusNotFound)
}

// TestRemovedParameterizedChannelReturnsNotFound verifies that an instance of a parameterized
// channel that was used before the reload is not served after the channel was removed.
func (suite *WebsocketReloadSuite) TestRemovedParameterizedChannelReturnsNotFound() {
	suite.reset()

	suite.writeSpec(reloadChatChannel)

	suite.waitForStatus("/rooms/1", http.StatusNotFound)
}

// TestOpenConnectionOfRemovedChannelIsClosed verifies that a client connected to a channel
// that is removed from the file is told that the channel is gone.
func (suite *WebsocketReloadSuite) TestOpenConnectionOfRemovedChannelIsClosed() {
	suite.reset()
	c := suite.connect("/chat")

	suite.writeSpec(reloadRoomsChannel)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	suite.Equal(websocket.StatusGoingAway, websocket.CloseStatus(err), "read error: %v", err)
	var closeErr websocket.CloseError
	suite.Require().ErrorAs(err, &closeErr)
	suite.Equal("channel removed", closeErr.Reason)
}
