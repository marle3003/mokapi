package kafka

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewBytes_Len(t *testing.T) {
	b := NewBytes([]byte("foobar"))
	require.Equal(t, 6, b.Size())
}

func TestNewBytes_Read(t *testing.T) {
	b := NewBytes([]byte("foobar"))
	result := [6]byte{}
	n, err := b.Read(result[:])
	require.NoError(t, err)
	require.Equal(t, 6, n)
	require.Equal(t, "foobar", string(result[:]))
}

func TestBytesToString(t *testing.T) {
	b := NewBytes([]byte("foobar"))
	require.Equal(t, "foobar", BytesToString(b))
}
