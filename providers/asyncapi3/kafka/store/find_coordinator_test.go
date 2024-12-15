package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
	"mokapi/kafka/kafkatest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"
)

func TestFindCoordinator(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"find group",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "127.0.0.1:9092")))
				r := kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				})
				r.Host = "127.0.0.1:9092"
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, r)

				res, ok := rr.Message.(*findCoordinator.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode, "expected no kafka error")

				require.Equal(t, "127.0.0.1", res.Host)
				require.Equal(t, int32(9092), res.Port)
			},
		},
		{
			"unsupported key type",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: 10,
				})
				s.ServeMessage(rr, r)
				res, ok := rr.Message.(*findCoordinator.Response)
				require.True(t, ok)
				require.Equal(t, kafka.UnknownServerError, res.ErrorCode)
				require.Equal(t, "unsupported request key_type=10", res.ErrorMessage)
			},
		},
		{
			"unknown broker",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				r := kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				})
				r.Host = "127.0.0.1:9092"
				s.ServeMessage(rr, r)
				res, ok := rr.Message.(*findCoordinator.Response)
				require.True(t, ok)
				require.Equal(t, kafka.UnknownServerError, res.ErrorCode)
				require.Equal(t, "broker 127.0.0.1:9092 not found", res.ErrorMessage)
			},
		},
		{
			"broker without host",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", ":9092")))

				r := kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				})
				r.Host = "127.0.0.1:9092"
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, r)

				res, ok := rr.Message.(*findCoordinator.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode, "expected no kafka error")

				require.Equal(t, "127.0.0.1", res.Host)
				require.Equal(t, int32(9092), res.Port)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapi3test.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}
