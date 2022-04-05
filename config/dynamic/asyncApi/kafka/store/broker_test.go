package store

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBroker_Addr(t *testing.T) {
	b := &Broker{Host: "bar", Port: 9092}
	require.Equal(t, "bar:9092", b.Addr())
}
