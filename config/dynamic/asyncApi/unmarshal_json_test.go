package asyncApi_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"testing"
)

func Test_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		target   any
		expected any
	}{
		{
			name:     "ParameterRef $ref",
			input:    `{"$ref":"foo"}`,
			target:   &asyncApi.ParameterRef{},
			expected: &asyncApi.ParameterRef{Reference: dynamic.Reference{Ref: "foo"}},
		},
		{
			name:     "ParameterRef value",
			input:    `{"location":"foo"}`,
			target:   &asyncApi.ParameterRef{},
			expected: &asyncApi.ParameterRef{Value: &asyncApi.Parameter{Location: "foo"}},
		},
		{
			name:     "MessageRef ref",
			input:    `{"$ref":"foo"}`,
			target:   &asyncApi.MessageRef{},
			expected: &asyncApi.MessageRef{Ref: "foo"},
		},
		{
			name:     "MessageRef value",
			input:    `{"messageId":"foo"}`,
			target:   &asyncApi.MessageRef{},
			expected: &asyncApi.MessageRef{Value: &asyncApi.Message{MessageId: "foo"}},
		},
		{
			name:     "ChannelRef ref",
			input:    `{"$ref":"foo"}`,
			target:   &asyncApi.ChannelRef{},
			expected: &asyncApi.ChannelRef{Ref: "foo"},
		},
		{
			name:   "ChannelRef value",
			input:  `{"description":"foo"}`,
			target: &asyncApi.ChannelRef{},
			expected: &asyncApi.ChannelRef{
				Value: &asyncApi.Channel{
					Description: "foo",
					Bindings: asyncApi.ChannelBindings{
						Kafka: asyncApi.TopicBindings{
							Partitions:            1,
							ValueSchemaValidation: true,
							KeySchemaValidation:   true,
						},
					},
				},
			},
		},
		{
			name:     "ServerRef ref",
			input:    `{"$ref":"foo"}`,
			target:   &asyncApi.ServerRef{},
			expected: &asyncApi.ServerRef{Ref: "foo"},
		},
		{
			name:   "ServerRef value",
			input:  `{"url":"foo"}`,
			target: &asyncApi.ServerRef{},
			expected: &asyncApi.ServerRef{
				Value: &asyncApi.Server{
					Url: "foo",
				},
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := json.Unmarshal([]byte(tc.input), &tc.target)
			require.NoError(t, err)
			require.Equal(t, tc.expected, tc.target)
		})
	}
}
