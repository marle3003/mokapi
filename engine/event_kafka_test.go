package engine_test

import (
	"mokapi/engine"
	"mokapi/engine/common"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent_Kafka(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "api filtering",
			test: func(t *testing.T) {
				d := &engine.KafkaEventDispatcher{}

				var called []string
				d.Register("foo", &engine.KafkaEventHandler{
					Filter: common.KafkaFilter{
						Api: "foo",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "foo")
						return true, nil
					},
				})
				d.Register("bar", &engine.KafkaEventHandler{
					Filter: common.KafkaFilter{
						Api: "bar",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "bar")
						return true, nil
					},
				})

				d.EmitKafka(&common.KafkaEventRecord{
					Api: "foo",
				})

				require.Equal(t, []string{"foo"}, called)
			},
		},
		{
			name: "topic filtering",
			test: func(t *testing.T) {
				d := &engine.KafkaEventDispatcher{}

				var called []string
				d.Register("foo", &engine.KafkaEventHandler{
					Filter: common.KafkaFilter{
						Topic: "foo",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "foo")
						return true, nil
					},
				})
				d.Register("bar", &engine.KafkaEventHandler{
					Filter: common.KafkaFilter{
						Topic: "bar",
					},
					Execute: func(ctx *common.EventContext) (bool, error) {
						called = append(called, "bar")
						return true, nil
					},
				})

				d.EmitKafka(&common.KafkaEventRecord{
					Topic: "foo",
				})

				require.Equal(t, []string{"foo"}, called)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
