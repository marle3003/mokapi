package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/kafkatest"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"
)

func TestHeartbeat(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, s *store.Store)
	}{
		{
			"not in group",
			func(t *testing.T, s *store.Store) {
				rr := kafkatest.NewRecorder()
				s.ServeMessage(rr, kafkatest.NewRequest("kafkatest", 3, &heartbeat.Request{
					GroupId:  "foo",
					MemberId: "bar",
				}))
				res, ok := rr.Message.(*heartbeat.Response)
				require.True(t, ok)
				require.Equal(t, kafka.UnknownMemberId, res.ErrorCode)
			},
		},
		{
			"group balancing",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithServer("", "kafka", b.Addr)))
				j, err := b.Client().JoinGroup(3, &joinGroup.Request{GroupId: "foo", MemberId: "bar"})
				require.NoError(t, err)
				require.Equal(t, kafka.None, j.ErrorCode)

				r, err := b.Client().Heartbeat(3, &heartbeat.Request{
					GroupId:  "foo",
					MemberId: "bar",
				})
				require.NoError(t, err)
				require.Equal(t, kafka.RebalanceInProgress, r.ErrorCode)
			},
		},
		{
			"synced",
			func(t *testing.T, s *store.Store) {
				b := kafkatest.NewBroker(kafkatest.WithHandler(s))
				defer b.Close()
				s.Update(asyncapi3test.NewConfig(asyncapi3test.WithServer("", "kafka", b.Addr)))
				err := b.Client().JoinSyncGroup("foo", "TestGroup", 3, 3)
				require.NoError(t, err)
				r, err := b.Client().Heartbeat(3, &heartbeat.Request{
					GroupId:         "TestGroup",
					GenerationId:    0,
					MemberId:        "foo",
					GroupInstanceId: "",
				})
				require.NoError(t, err)
				require.Equal(t, kafka.None, r.ErrorCode)
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
