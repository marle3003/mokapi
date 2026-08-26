package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOperation(t *testing.T) {
	s := `
asyncapi: "3.0.0"
channels:
  userSignedUp:
    messages:
      userSignedUp:
        payload:
          type: string

operations:
  userSignupOperation:
    action: send
    channel:
      $ref: '#/channels/userSignedUp'
    messages:
      - $ref: '#/channels/userSignedUp/messages/userSignedUp'
`
	var cfg *asyncapi3.Config
	err := yaml.Unmarshal([]byte(s), &cfg)
	require.NoError(t, err)
	err = cfg.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: cfg}, &dynamictest.Reader{})
	require.NoError(t, err)

	require.Len(t, cfg.Operations, 1)
	op := cfg.Operations["userSignupOperation"]
	require.NotNil(t, op)
	require.NotNil(t, op.Value)
	require.Equal(t, "send", op.Value.Action)
	require.Equal(t, "userSignedUp", op.Value.Channel.Value.Name)
	require.IsType(t, &schema.Schema{}, op.Value.Messages[0].Value.Payload.Value)
	js := op.Value.Messages[0].Value.Payload.Value.(*schema.Schema)
	require.Equal(t, "string", js.Type.String())
}

func TestOperation_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.OperationRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.OperationRef{
				Reference: dynamic.Reference[*asyncapi3.OperationRef]{
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
			s: &asyncapi3.OperationRef{
				Value: &asyncapi3.Operation{
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

func TestOperationTrait_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.OperationTraitRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.OperationTraitRef{
				Reference: dynamic.Reference[*asyncapi3.OperationTraitRef]{
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
			s: &asyncapi3.OperationTraitRef{
				Value: &asyncapi3.OperationTrait{
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
