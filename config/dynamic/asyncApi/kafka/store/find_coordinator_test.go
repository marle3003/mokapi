package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/kafka"
	"mokapi/kafka/findCoordinator"
	"mokapi/kafka/kafkatest"
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
				s.Update(asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092")))
				r := kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				})
				r.Host = "127.0.0.1"
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
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &findCoordinator.Request{
					Key:     "foo",
					KeyType: findCoordinator.KeyTypeGroup,
				}))

				res, ok := rr.Message.(*findCoordinator.Response)
				require.True(t, ok)
				require.Equal(t, kafka.UnknownServerError, res.ErrorCode)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t, store.New(asyncapitest.NewConfig()))
		})
	}
}
