package acceptance

import (
	"context"
	"encoding/json"
	"mokapi/config/static"
	"mokapi/try"
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
