package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
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

func TestChannel_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.ChannelRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.ChannelRef{
				Reference: dynamic.Reference[*asyncapi3.ChannelRef]{
					Ref: "/foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"/foo"}`, json)
				require.Equal(t, "$ref: /foo\n", yaml)
			},
		},
		{
			name: "value",
			s: &asyncapi3.ChannelRef{
				Value: &asyncapi3.Channel{
					Description: "foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"description":"foo"}`, json)
				require.Equal(t, "description: foo\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}

func TestChannel_GetMessageByPayload_EmptyEntries(t *testing.T) {
	var cfg *asyncapi3.Config
	require.NoError(t, yaml.Unmarshal([]byte(`
asyncapi: 3.0.0
info: {title: foo, version: '1.0'}
channels:
  orders:
    messages:
      empty:
      order:
        name: order
operations:
  empty:
  noChannel:
    action: send
  send:
    action: send
    channel:
      $ref: '#/channels/orders'
`), &cfg))
	c := &dynamic.Config{Data: cfg}
	require.NoError(t, cfg.Parse(c, &dynamictest.Reader{}))

	m, err := cfg.Channels["orders"].Value.GetMessageByPayload(nil)
	require.NoError(t, err)
	require.Equal(t, "order", m.Name)
}
