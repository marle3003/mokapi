package websocket_test

import (
	"encoding/json"
	"mokapi/engine/common"
	"mokapi/providers/asyncapi3/websocket"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageLog_Actions(t *testing.T) {
	l := &websocket.MessageLog{}
	require.Empty(t, l.GetActions())

	actions := []*common.Action{{Duration: 1}}
	l.SetActions(actions)
	require.Equal(t, actions, l.GetActions())
}

func TestConnectLog_Actions(t *testing.T) {
	l := &websocket.ConnectLog{}
	require.Empty(t, l.GetActions())

	actions := []*common.Action{{Duration: 1}}
	l.SetActions(actions)
	require.Equal(t, actions, l.GetActions())
}

func TestMessageLog_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		log  func() *websocket.MessageLog
		json string
	}{
		{
			name: "without actions",
			log: func() *websocket.MessageLog {
				return &websocket.MessageLog{Channel: "chat", MessageId: "msg", Direction: websocket.Send, Api: "foo"}
			},
			json: `{"channel":"chat","message":null,"messageId":"msg","direction":"send","client":{"id":"","query":null,"header":null,"address":"","server":""},"api":"foo","actions":null}`,
		},
		{
			name: "with actions and validation errors",
			log: func() *websocket.MessageLog {
				l := &websocket.MessageLog{
					Channel:          "chat",
					Message:          &websocket.LogValue{Value: "hi"},
					ValidationErrors: map[string]string{"m": "err"},
				}
				l.SetActions([]*common.Action{{Duration: 2}})
				return l
			},
			json: `{"channel":"chat","message":{"value":"hi","binary":null},"messageId":"","direction":"","client":{"id":"","query":null,"header":null,"address":"","server":""},"api":"","actions":[{"duration":2,"tags":null,"parameters":null,"logs":null,"error":null}],"validationErrors":{"m":"err"}}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.log())
			require.NoError(t, err)
			require.Equal(t, tc.json, string(b))
		})
	}
}

func TestConnectLog_MarshalJSON(t *testing.T) {
	l := &websocket.ConnectLog{Type: "connect", Channel: "chat", Api: "foo"}
	l.SetActions([]*common.Action{{Duration: 2}})

	b, err := json.Marshal(l)
	require.NoError(t, err)
	require.Equal(t, `{"type":"connect","channel":"chat","client":{"id":"","query":null,"header":null,"address":"","server":""},"api":"foo","actions":[{"duration":2,"tags":null,"parameters":null,"logs":null,"error":null}]}`, string(b))
}

func TestLog_ConcurrentSetActionsAndMarshalJSON(t *testing.T) {
	message := &websocket.MessageLog{}
	connect := &websocket.ConnectLog{}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			message.SetActions([]*common.Action{{Duration: 1}})
			connect.SetActions([]*common.Action{{Duration: 1}})
		}()
		go func() {
			defer wg.Done()
			_, _ = json.Marshal(message)
			_, _ = json.Marshal(connect)
		}()
	}
	wg.Wait()

	require.Len(t, message.GetActions(), 1)
	require.Len(t, connect.GetActions(), 1)
}
