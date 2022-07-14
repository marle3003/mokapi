package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/kafka"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/listgroup"
	"testing"
)

func TestListGroup(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"empty",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 4,
					&listgroup.Request{},
				))
				res, ok := rr.Message.(*listgroup.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Groups, 0)
			},
		},
		{
			"with group state",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithServer("", "kafka", "")))
				group := s.GetOrCreateGroup("foo", 0)
				group.State = store.Joining
				g := group.NewGeneration()
				g.Members[""] = &store.Member{}

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 4,
					&listgroup.Request{},
				))
				res, ok := rr.Message.(*listgroup.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Groups, 1)
				require.Equal(t, "PreparingRebalance", res.Groups[0].GroupState)
			},
		},
		{
			"filtering",
			func(t *testing.T, s *store.Store) {
				s.Update(asyncapitest.NewConfig(asyncapitest.WithServer("", "kafka", "")))
				s.GetOrCreateGroup("foo", 0)
				group := s.GetOrCreateGroup("bar", 0)
				group.State = store.AwaitingSync
				g := group.NewGeneration()
				g.Members[""] = &store.Member{}

				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 4,
					&listgroup.Request{StatesFilter: []string{"Empty"}},
				))
				res, ok := rr.Message.(*listgroup.Response)
				require.True(t, ok)
				require.Equal(t, kafka.None, res.ErrorCode)
				require.Len(t, res.Groups, 1)
				require.Equal(t, "foo", res.Groups[0].GroupId)
				require.Equal(t, "Empty", res.Groups[0].GroupState)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := store.New(asyncapitest.NewConfig())
			defer s.Close()
			tc.fn(t, s)
		})
	}
}
