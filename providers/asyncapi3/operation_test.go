package asyncapi3_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	json "mokapi/schema/json/schema"
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
	require.IsType(t, &json.Schema{}, op.Value.Messages[0].Value.Payload.Value)
	js := op.Value.Messages[0].Value.Payload.Value.(*json.Schema)
	require.Equal(t, "string", js.Type.String())
}
