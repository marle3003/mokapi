package store

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/engine/enginetest"
	"testing"
)

func TestStore(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			"empty",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 0, len(s.Brokers()))
				require.Equal(t, 0, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				require.Nil(t, s.Topic("foo"), "topic not exists")
			},
		},
		{
			"server",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(
					asyncapitest.WithServer("foo", "kafka", "foo:9092"),
				), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 1, len(s.Brokers()))
				require.Equal(t, 0, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				b, ok := s.Broker(0)
				require.Equal(t, true, ok)
				require.Equal(t, "foo", b.Name)
			},
		},
		{
			"topic",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(
					asyncapitest.WithChannel("foo"),
				), enginetest.NewEngine())
				defer s.Close()
				require.Equal(t, 0, len(s.Brokers()))
				require.Equal(t, 1, len(s.Topics()))
				require.Equal(t, 0, len(s.Groups()))
				topic := s.Topic("foo")
				require.NotNil(t, topic, "topic is not nil")
				require.Equal(t, "foo", topic.Name)
				require.Len(t, topic.Partitions, 1)
			},
		},
		{
			"create topic",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(), enginetest.NewEngine())
				defer s.Close()
				topic, err := s.NewTopic("foo", asyncapitest.NewChannel())
				require.NoError(t, err)
				require.Equal(t, "foo", topic.Name)
				require.Equal(t, 1, len(topic.Partitions))
			},
		},
		{
			"create topic, already exists",
			func(t *testing.T) {
				s := New(asyncapitest.NewConfig(asyncapitest.WithChannel("foo")), enginetest.NewEngine())
				defer s.Close()
				_, err := s.NewTopic("foo", asyncapitest.NewChannel())
				require.Error(t, err, "topic foo already exists")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}
