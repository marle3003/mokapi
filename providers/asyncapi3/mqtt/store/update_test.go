package store_test

import (
	"mokapi/engine/enginetest"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schema/schematest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func newUpdateTestStore(t *testing.T, cfg *asyncapi3.Config) *store.Store {
	s := store.New(cfg, enginetest.NewEngine(), &eventstest.Handler{}, monitor.NewMqtt())
	t.Cleanup(s.Close)
	return s
}

func mqttConfig(opts ...asyncapi3test.ConfigOptions) *asyncapi3.Config {
	return asyncapi3test.NewConfig(append([]asyncapi3test.ConfigOptions{asyncapi3test.WithInfo("test-server", "", "")}, opts...)...)
}

func TestStore_Update(t *testing.T) {
	t.Run("static topic keeps retained message", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(
			asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg"))))

		_, err := s.Publish(&mqtt.PublishRequest{Topic: "foo", Data: []byte("hello")}, store.PublishArgs{Retain: true})
		require.NoError(t, err)

		s.Update(mqttConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg"))))

		topic, ok := s.Topic("foo")
		require.True(t, ok)
		require.NotNil(t, topic.Retained, "retained message must survive a reload")
		require.Equal(t, "hello", string(topic.Retained.Data))
	})

	t.Run("static topic uses new channel", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg"))))
		ch := asyncapi3test.NewChannel(asyncapi3test.WithMessage("msg", asyncapi3test.WithPayload(schematest.New("integer"))))

		s.Update(mqttConfig(asyncapi3test.AddChannel("foo", ch)))

		topic, ok := s.Topic("foo")
		require.True(t, ok)
		require.Same(t, ch, topic.Config())
	})

	t.Run("parameterized topic validates against new channel", func(t *testing.T) {
		withPayload := func(typ string) *asyncapi3.Config {
			return mqttConfig(asyncapi3test.WithChannel("sensors/{id}/data",
				asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
				asyncapi3test.WithMessage("msg", asyncapi3test.WithPayload(schematest.New(typ)))))
		}
		s := newUpdateTestStore(t, withPayload("string"))

		_, err := s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{})
		require.NoError(t, err)

		s.Update(withPayload("integer"))

		reason, err := s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{})
		require.Error(t, err, "payload must be validated against the new channel")
		require.Equal(t, mqtt.PayloadFormatInvalid, reason)
	})

	t.Run("parameterized topic is removed with its channel", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(asyncapi3test.WithChannel("sensors/{id}/data",
			asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("msg"))))

		_, err := s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{Retain: true})
		require.NoError(t, err)

		s.Update(mqttConfig(asyncapi3test.WithChannel("other", asyncapi3test.WithMessage("msg"))))

		require.NotContains(t, topicNames(s.Topics()), "sensors/1/data")
		_, ok := s.Topic("sensors/1/data")
		require.False(t, ok)
	})

	t.Run("topic of removed channel is removed", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(
			asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg")),
			asyncapi3test.WithChannel("bar", asyncapi3test.WithMessage("msg"))))

		s.Update(mqttConfig(asyncapi3test.WithChannel("bar", asyncapi3test.WithMessage("msg"))))

		names := topicNames(s.Topics())
		require.NotContains(t, names, "foo")
		require.Contains(t, names, "bar")
		require.Contains(t, names, "$SYS/broker/version", "system topics must stay")
	})

	t.Run("$SYS topics stay", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg"))))

		s.Update(mqttConfig())

		topic, ok := s.Topic("$SYS/broker/version")
		require.True(t, ok)
		require.NotNil(t, topic.Retained)
		require.Contains(t, topicNames(s.Topics()), "$SYS/broker/uptime")
	})

	t.Run("topic moves from static to parameterized channel", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(asyncapi3test.WithChannel("sensors/1/data", asyncapi3test.WithMessage("msg"))))
		_, err := s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{Retain: true})
		require.NoError(t, err)

		ch := asyncapi3test.NewChannel(
			asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("msg"))
		s.Update(mqttConfig(asyncapi3test.AddChannel("sensors/{id}/data", ch)))

		topic, ok := s.Topic("sensors/1/data")
		require.True(t, ok)
		require.Same(t, ch, topic.Config())
		require.Equal(t, "hello", string(topic.Retained.Data))
	})

	t.Run("topic moves from parameterized to static channel", func(t *testing.T) {
		s := newUpdateTestStore(t, mqttConfig(asyncapi3test.WithChannel("sensors/{id}/data",
			asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
			asyncapi3test.WithMessage("msg"))))
		_, err := s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{Retain: true})
		require.NoError(t, err)

		s.Update(mqttConfig(asyncapi3test.WithChannel("sensors/1/data", asyncapi3test.WithMessage("msg"))))

		topic, ok := s.Topic("sensors/1/data")
		require.True(t, ok)
		require.Nil(t, topic.Config().Parameters)
		require.Equal(t, "hello", string(topic.Retained.Data))
	})

	t.Run("update while publishing", func(t *testing.T) {
		newConfig := func() *asyncapi3.Config {
			return mqttConfig(
				asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("msg")),
				asyncapi3test.WithChannel("sensors/{id}/data",
					asyncapi3test.WithParameter("id", &asyncapi3.Parameter{}),
					asyncapi3test.WithMessage("msg")))
		}
		s := newUpdateTestStore(t, newConfig())

		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				s.Update(newConfig())
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_, _ = s.Publish(&mqtt.PublishRequest{Topic: "foo", Data: []byte("hello")}, store.PublishArgs{Retain: true})
				_, _ = s.Publish(&mqtt.PublishRequest{Topic: "sensors/1/data", Data: []byte("hello")}, store.PublishArgs{Retain: true})
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				for _, topic := range s.Topics() {
					_ = topic.Config()
				}
			}
		}()
		wg.Wait()

		topic, ok := s.Topic("foo")
		require.True(t, ok)
		require.NotNil(t, topic.Retained)

		topic, ok = s.Topic("sensors/1/data")
		require.True(t, ok)
		require.NotNil(t, topic.Retained)
	})
}
