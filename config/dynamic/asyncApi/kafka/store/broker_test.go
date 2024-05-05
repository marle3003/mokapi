package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"testing"
)

func TestBroker_Addr(t *testing.T) {
	b := &store.Broker{Host: "bar", Port: 9092}
	require.Equal(t, "bar:9092", b.Addr())
}
