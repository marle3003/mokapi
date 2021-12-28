package store

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/test"
	"testing"
)

func TestStore(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			"empty",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig())
				test.Equals(t, 0, len(s.Brokers()))
				test.Equals(t, 0, len(s.Topics()))
				test.Equals(t, 0, len(s.Groups()))
				test.Assert(t, s.Topic("foo") == nil, "topic not exists")
			},
		},
		{
			"server",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(
					asyncapitest.WithServer("foo", "kafka", "foo:9092"),
				))
				test.Equals(t, 1, len(s.Brokers()))
				test.Equals(t, 0, len(s.Topics()))
				test.Equals(t, 0, len(s.Groups()))
				b, ok := s.Broker(0)
				test.Equals(t, true, ok)
				test.Equals(t, "foo", b.Name())
			},
		},
		{
			"topic",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo"),
				))
				test.Equals(t, 0, len(s.Brokers()))
				test.Equals(t, 1, len(s.Topics()))
				test.Equals(t, 0, len(s.Groups()))
				topic := s.Topic("foo")
				test.Assert(t, topic != nil, "topic is not nil")
				test.Equals(t, "foo", topic.Name())
				test.Equals(t, 1, len(topic.Partitions()))
			},
		},
		{
			"create topic",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig())
				topic, err := s.NewTopic("foo", asyncapitest.NewChannel())
				test.Ok(t, err)
				test.Equals(t, "foo", topic.Name())
				test.Equals(t, 1, len(topic.Partitions()))
			},
		},
		{
			"create topic, already exists",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")))
				_, err := s.NewTopic("foo", asyncapitest.NewChannel())
				test.EqualError(t, "topic foo already exists", err)
			},
		},
	}

	t.Parallel()
	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			d.fn(t)
		})
	}
}
