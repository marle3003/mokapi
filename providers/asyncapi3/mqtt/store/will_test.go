package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schema/schematest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWillMessage(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store, eh *eventstest.Handler)
	}{
		{
			name: "will is delivered to subscriber but not retained when retain flag is not set",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte("bar"), false)
				observer := newClient("observer", s)
				defer observer.close()
				observer.connect()
				subscribe(observer, "foo")

				disconnectWithWill(t, device)

				pub := receive(t, observer)
				require.Equal(t, "foo", pub.Topic)
				require.Equal(t, []byte("bar"), pub.Data)

				topic, ok := s.Topic("foo")
				require.True(t, ok)
				require.Nil(t, topic.Retained)

				late := newClient("late", s)
				defer late.close()
				late.connect()
				subscribe(late, "foo")
				requireNoMessage(t, late)
			},
		},
		{
			name: "will with retain flag is retained",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte("bar"), true)

				disconnectWithWill(t, device)

				topic, ok := s.Topic("foo")
				require.True(t, ok)
				require.NotNil(t, topic.Retained)
				require.Equal(t, []byte("bar"), topic.Retained.Data)

				late := newClient("late", s)
				defer late.close()
				late.connect()
				subscribe(late, "foo")
				pub := receive(t, late)
				require.Equal(t, "foo", pub.Topic)
				require.Equal(t, []byte("bar"), pub.Data)
			},
		},
		{
			name: "retained will with empty payload removes the retained message",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				publisher := newClient("publisher", s)
				defer publisher.close()
				publisher.connect()
				publishRetained(publisher, "foo", []byte("old"))
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte{}, true)

				disconnectWithWill(t, device)

				topic, ok := s.Topic("foo")
				require.True(t, ok)
				require.Nil(t, topic.Retained)
			},
		},
		{
			name: "will without retain flag keeps an existing retained message",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				publisher := newClient("publisher", s)
				defer publisher.close()
				publisher.connect()
				publishRetained(publisher, "foo", []byte("old"))
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte("bar"), false)

				disconnectWithWill(t, device)

				topic, ok := s.Topic("foo")
				require.True(t, ok)
				require.NotNil(t, topic.Retained)
				require.Equal(t, []byte("old"), topic.Retained.Data)
			},
		},
		{
			name: "will on topic with parameter is accepted and delivered",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithChannel("sensors/{sensorId}/status",
						asyncapi3test.WithParameter("sensorId", &asyncapi3.Parameter{}),
					),
				))
				device := newClient("device", s)
				defer device.close()
				rr := device.send(&mqtt.Message{
					Payload: &mqtt.ConnectRequest{
						ClientId: device.clientCtx.ClientId,
						WillFlag: true,
						Topic:    "sensors/1234z/status",
						Message:  []byte("offline"),
					},
					Context: device.ctx,
				})
				require.Equal(t, mqtt.Success, rr.Message.Payload.(*mqtt.ConnectResponse).ReasonCode)
				observer := newClient("observer", s)
				defer observer.close()
				observer.connect()
				subscribe(observer, "sensors/1234z/status")

				disconnectWithWill(t, device)

				pub := receive(t, observer)
				require.Equal(t, "sensors/1234z/status", pub.Topic)
				require.Equal(t, []byte("offline"), pub.Data)
			},
		},
		{
			name: "will with invalid payload is neither delivered nor retained",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				s.Update(asyncapi3test.NewConfig(
					asyncapi3test.WithInfo("test-server", "", ""),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("bar", asyncapi3test.WithPayload(schematest.New("integer"))),
					),
				))
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte("not a number"), true)
				observer := newClient("observer", s)
				defer observer.close()
				observer.connect()
				subscribe(observer, "foo")

				disconnectWithWill(t, device)

				requireNoMessage(t, observer)
				topic, ok := s.Topic("foo")
				require.True(t, ok)
				require.Nil(t, topic.Retained)
				require.Empty(t, eh.GetEvents(events.NewTraits().WithNamespace("mqtt").With("type", "message")))
			},
		},
		{
			name: "will is logged as message of the disconnected client",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				device := newClient("device", s)
				defer device.close()
				connectWithWill(device, "foo", []byte("bar"), true)

				disconnectWithWill(t, device)

				evts := eh.GetEvents(events.NewTraits().WithNamespace("mqtt").With("type", "message"))
				require.Len(t, evts, 1)
				msg := evts[0].Data.(*store.LogMessage)
				require.Equal(t, "foo", msg.Topic)
				require.Equal(t, "device", msg.ClientId)
				require.True(t, msg.Retain)
				require.Equal(t, "bar", msg.Message.Value)
			},
		},
		{
			name: "disconnect with will reason but without will does not panic",
			test: func(t *testing.T, s *store.Store, eh *eventstest.Handler) {
				device := newClient("device", s)
				defer device.close()
				device.connect()

				disconnectWithWill(t, device)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			eh := &eventstest.Handler{}
			s := store.New(
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("test-server", "", ""), asyncapi3test.WithChannel("foo")),
				enginetest.NewEngine(),
				eh,
				monitor.NewMqtt(),
			)
			defer s.Close()

			tc.test(t, s, eh)
		})
	}
}

func connectWithWill(c *client, topic string, data []byte, retain bool) {
	c.send(&mqtt.Message{
		Payload: &mqtt.ConnectRequest{
			ClientId:   c.clientCtx.ClientId,
			WillFlag:   true,
			WillRetain: retain,
			Topic:      topic,
			Message:    data,
		},
		Context: c.ctx,
	})
}

// disconnectWithWill turns a panic of the store into a test failure, so that
// the remaining cases still run.
func disconnectWithWill(t *testing.T, c *client) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("disconnect panicked: %v", r)
		}
	}()
	c.send(&mqtt.Message{
		Payload: &mqtt.DisconnectRequest{Reason: mqtt.DisconnectWithWillMessage},
		Context: c.ctx,
	})
}

func receive(t *testing.T, c *client) *mqtt.PublishRequest {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Second))
	res := &mqtt.Message{}
	require.NoError(t, res.Read(c.conn, c.clientCtx))
	return res.Payload.(*mqtt.PublishRequest)
}

func requireNoMessage(t *testing.T, c *client) {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	res := &mqtt.Message{}
	err := res.Read(c.conn, c.clientCtx)
	require.True(t, os.IsTimeout(err), "no message expected, got: %v", err)
}
