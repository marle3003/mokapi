package store_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// handlerEmitter behaves like a JavaScript handler: it runs synchronously inside
// Publish and may publish to another topic itself.
type handlerEmitter struct {
	common.EventEmitter
	onMqtt func(evt *common.MqttMessageEvent)
}

func (e *handlerEmitter) EmitMqtt(evt *common.MqttMessageEvent) []*common.Action {
	e.onMqtt(evt)
	return nil
}

func newStoreWithEmitter(emitter common.EventEmitter, topics ...string) *store.Store {
	opts := []asyncapi3test.ConfigOptions{asyncapi3test.WithInfo("test-server", "", "")}
	for _, name := range topics {
		opts = append(opts, asyncapi3test.WithChannel(name, asyncapi3test.WithMessage("msg")))
	}
	return store.New(asyncapi3test.NewConfig(opts...), emitter, &eventstest.Handler{}, monitor.NewMqtt())
}

func TestPublish_HandlerPublishesToAnotherTopicWhileRequestIsServed(t *testing.T) {
	var s *store.Store
	var handlerErr error
	emitter := &handlerEmitter{onMqtt: func(evt *common.MqttMessageEvent) {
		if evt.Topic != "first" {
			return
		}
		_, handlerErr = s.Publish(
			&mqtt.PublishRequest{Topic: "second", Data: []byte("forwarded")},
			store.PublishArgs{Retain: true},
		)
	}}
	s = newStoreWithEmitter(emitter, "first", "second")
	defer s.Close()

	publisher := newClient("publisher", s)
	defer publisher.close()
	publisher.connect()

	done := make(chan *mqtt.Message)
	go func() {
		rr := publisher.send(&mqtt.Message{
			Header:  &mqtt.Header{QoS: 1},
			Payload: &mqtt.PublishRequest{MessageId: 1, Topic: "first", Data: []byte("hello")},
			Context: publisher.ctx,
		})
		done <- rr.Message
	}()

	select {
	case msg := <-done:
		require.NoError(t, handlerErr)
		require.Equal(t, mqtt.PublishSuccess, msg.Payload.(*mqtt.PublishResponse).ReasonCode)
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: publish from within a handler did not return")
	}

	second, ok := s.Topic("second")
	require.True(t, ok)
	require.Equal(t, "forwarded", string(second.Retained.Data))
}

func TestPublish_RetainsMessageOnTopicReplacedByReloadDuringHandler(t *testing.T) {
	var s *store.Store
	emitter := &handlerEmitter{onMqtt: func(*common.MqttMessageEvent) {
		// a reload replaces every topic while the handler runs
		s.Update(asyncapi3test.NewConfig(
			asyncapi3test.WithInfo("test-server", "", ""),
			asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg"))))
	}}
	s = newStoreWithEmitter(emitter, "foo")
	defer s.Close()

	_, err := s.Publish(&mqtt.PublishRequest{Topic: "foo", Data: []byte("hello")}, store.PublishArgs{Retain: true})
	require.NoError(t, err)

	topic, ok := s.Topic("foo")
	require.True(t, ok)
	require.NotNil(t, topic.Retained)
	require.Equal(t, "hello", string(topic.Retained.Data))
}

func TestStore_ConcurrentConnectSubscribePublish(t *testing.T) {
	s := newStoreWithEmitter(&handlerEmitter{onMqtt: func(*common.MqttMessageEvent) {}}, "foo")
	defer s.Close()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := newClient("client-"+strconv.Itoa(i), s)
			defer c.close()
			// a subscriber's connection must be read, otherwise net.Pipe blocks the sender
			go func() { _, _ = io.Copy(io.Discard, c.conn) }()

			c.connect()
			c.send(&mqtt.Message{
				Payload: &mqtt.SubscribeRequest{
					MessageId: 1,
					Topics:    []mqtt.SubscribeTopic{{Name: "foo", QoS: 1}},
				},
				Context: c.ctx,
			})
			for j := 0; j < 5; j++ {
				c.send(&mqtt.Message{
					Header:  &mqtt.Header{QoS: 1, Retain: true},
					Payload: &mqtt.PublishRequest{MessageId: uint16(j + 1), Topic: "foo", Data: []byte("bar")},
					Context: c.ctx,
				})
				c.send(&mqtt.Message{Payload: &mqtt.PingRequest{}, Context: c.ctx})
				// reconnect with the same client id replaces the client context
				c.connect()
				_ = s.Clients()
			}
		}()
	}
	wg.Wait()
}

func TestStore_EmptyChannelNextToValidChannel(t *testing.T) {
	cfg := asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg")),
	)
	cfg.Channels["empty"] = nil
	require.NoError(t, cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{}))

	var s *store.Store
	require.NotPanics(t, func() {
		s = store.New(cfg, &handlerEmitter{}, &eventstest.Handler{}, monitor.NewMqtt())
	})
	defer s.Close()

	require.Contains(t, s.Topics, "foo")
	require.NotContains(t, s.Topics, "empty")
}
