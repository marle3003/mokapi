package try

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port := GetFreePort()
	require.Greater(t, port, 0)
}

func TestMustUrl(t *testing.T) {
	u := MustUrl("http://localhost")
	require.NotNil(t, u)

	require.Panics(t, func() { MustUrl(":") })
}
