package asyncapi3_test

import (
	"encoding/json"
	"mokapi/providers/asyncapi3"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestChannel_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *asyncapi3.Channel, err error)
	}{
		{
			name: "tags",
			data: `{ "tags": [ { "name": "foo", "description": "bar" } ] }`,
			test: func(t *testing.T, cfg *asyncapi3.Channel, err error) {
				require.NoError(t, err)
				require.Len(t, cfg.Tags, 1)
				require.Equal(t, "foo", cfg.Tags[0].Value.Name)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ch *asyncapi3.Channel
			err := json.Unmarshal([]byte(tc.data), &ch)
			tc.test(t, ch, err)
		})
	}
}

func TestChannel_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *asyncapi3.Channel, err error)
	}{
		{
			name: "tags",
			data: `
tags:
  - name: foo
    description: bar
`,
			test: func(t *testing.T, cfg *asyncapi3.Channel, err error) {
				require.NoError(t, err)
				require.Len(t, cfg.Tags, 1)
				require.Equal(t, "foo", cfg.Tags[0].Value.Name)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ch *asyncapi3.Channel
			err := yaml.Unmarshal([]byte(tc.data), &ch)
			tc.test(t, ch, err)
		})
	}
}
