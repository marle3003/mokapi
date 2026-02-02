package store_test

import (
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKafkaLog_Title(t *testing.T) {
	v := store.KafkaMessageLog{}
	require.Equal(t, "", v.Title())

	v = store.KafkaMessageLog{Key: store.LogValue{Value: "foo"}}
	require.Equal(t, "foo", v.Title())

	v = store.KafkaMessageLog{Key: store.LogValue{Binary: []byte("foo")}}
	require.Equal(t, "foo", v.Title())
}
