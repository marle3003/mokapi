package engine

import (
	"errors"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectMessage(t *testing.T) {
	created := asyncapi3test.NewMessage(asyncapi3test.WithMessageName("created"))
	deleted := asyncapi3test.NewMessage(asyncapi3test.WithMessageName("deleted"))
	ch := asyncapi3test.NewChannel(
		asyncapi3test.WithChannelAddress("orders"),
		asyncapi3test.UseMessage("orderCreated", &asyncapi3.MessageRef{Value: created}),
		asyncapi3test.UseMessage("orderDeleted", &asyncapi3.MessageRef{Value: deleted}),
	)

	testcases := []struct {
		name      string
		messageId string
		pick      func() (*asyncapi3.Message, error)
		test      func(t *testing.T, msg *asyncapi3.Message, err error)
	}{
		{
			name:      "messageId selects the message and skips the random choice",
			messageId: "orderDeleted",
			pick: func() (*asyncapi3.Message, error) {
				require.Fail(t, "random choice must not be used")
				return nil, nil
			},
			test: func(t *testing.T, msg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Same(t, deleted, msg)
			},
		},
		{
			name:      "unknown messageId is an error without fallback",
			messageId: "orderDeletd",
			pick: func() (*asyncapi3.Message, error) {
				require.Fail(t, "random choice must not be used")
				return nil, nil
			},
			test: func(t *testing.T, msg *asyncapi3.Message, err error) {
				require.EqualError(t, err, "message 'orderDeletd' not found in channel 'orders'")
				require.Nil(t, msg)
			},
		},
		{
			name: "without messageId the protocol's choice is used",
			pick: func() (*asyncapi3.Message, error) { return created, nil },
			test: func(t *testing.T, msg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Same(t, created, msg)
			},
		},
		{
			name: "without messageId the error of the protocol's choice is returned",
			pick: func() (*asyncapi3.Message, error) { return nil, errors.New("no message") },
			test: func(t *testing.T, msg *asyncapi3.Message, err error) {
				require.EqualError(t, err, "no message")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := selectMessage(ch, tc.messageId, tc.pick)
			tc.test(t, msg, err)
		})
	}
}

func TestGeneratePayload(t *testing.T) {
	withPayload := asyncapi3test.NewMessage(asyncapi3test.WithPayload(schematest.New("string", schematest.WithEnum([]any{"foo"}))))

	t.Run("generated from payload schema", func(t *testing.T) {
		v, err := generatePayload(withPayload)

		require.NoError(t, err)
		require.Equal(t, "foo", v)
	})
	t.Run("message without payload is null", func(t *testing.T) {
		v, err := generatePayload(asyncapi3test.NewMessage())

		require.NoError(t, err)
		require.Nil(t, v)
	})
	t.Run("no message is null", func(t *testing.T) {
		v, err := generatePayload(nil)

		require.NoError(t, err)
		require.Nil(t, v)
	})
	t.Run("any payload is generated from payload schema", func(t *testing.T) {
		v, err := generateAnyPayload(withPayload)

		require.NoError(t, err)
		require.Equal(t, "foo", v)
	})
	t.Run("any payload is generated for a message without payload schema", func(t *testing.T) {
		v, err := generateAnyPayload(asyncapi3test.NewMessage())

		require.NoError(t, err)
		require.NotNil(t, v)
	})
}

func TestAnyMessage(t *testing.T) {
	msg := asyncapi3test.NewMessage()

	require.Same(t, msg, anyMessage(asyncapi3test.NewChannel(
		asyncapi3test.UseMessage("empty", nil),
		asyncapi3test.UseMessage("unresolved", &asyncapi3.MessageRef{}),
		asyncapi3test.UseMessage("msg", &asyncapi3.MessageRef{Value: msg}),
	)))
	require.Nil(t, anyMessage(asyncapi3test.NewChannel()))
}

func TestEncodePayload(t *testing.T) {
	testcases := []struct {
		name string
		msg  *asyncapi3.Message
		data any
		test func(t *testing.T, b []byte, err error)
	}{
		{
			name: "valid data is encoded with the message's content type",
			msg: asyncapi3test.NewMessage(
				asyncapi3test.WithContentType("application/json"),
				asyncapi3test.WithPayload(schematest.New("object", schematest.WithProperty("id", schematest.New("integer")))),
			),
			data: map[string]any{"id": 1},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"id":1}`, string(b))
			},
		},
		{
			name: "invalid data is rejected",
			msg: asyncapi3test.NewMessage(
				asyncapi3test.WithContentType("application/json"),
				asyncapi3test.WithPayload(schematest.New("object", schematest.WithProperty("id", schematest.New("integer")))),
			),
			data: map[string]any{"id": "foo"},
			test: func(t *testing.T, b []byte, err error) {
				require.Error(t, err)
				require.Nil(t, b)
			},
		},
		{
			name: "message without payload encodes data as JSON",
			msg:  asyncapi3test.NewMessage(),
			data: map[string]any{"id": 1},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"id":1}`, string(b))
			},
		},
		{
			name: "no message encodes data as JSON",
			data: []any{1, 2},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `[1,2]`, string(b))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := encodePayload(tc.msg, tc.data)
			tc.test(t, b, err)
		})
	}
}
