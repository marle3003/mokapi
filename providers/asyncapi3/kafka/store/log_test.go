package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/asyncapi3/kafka/store"
	"testing"
)

func TestKafkaLog_Title(t *testing.T) {
	v := store.KafkaLog{}
	require.Equal(t, "", v.Title())

	v = store.KafkaLog{Key: store.LogValue{Value: "foo"}}
	require.Equal(t, "foo", v.Title())

	v = store.KafkaLog{Key: store.LogValue{Binary: []byte("foo")}}
	require.Equal(t, "foo", v.Title())
}
