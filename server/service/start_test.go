package service_test

import (
	"fmt"
	"mokapi/try"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type starter interface {
	Start() error
	Stop()
}

// startOnFreePort retries because another process may take the port between GetFreePort and Start.
func startOnFreePort[T starter](t *testing.T, create func(port string) T) (string, T) {
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		port := fmt.Sprintf("%v", try.GetFreePort())
		s := create(port)
		if err = s.Start(); err == nil {
			return port, s
		}
	}
	require.NoError(t, err)
	var zero T
	return "", zero
}

// requirePortIsBoundWhenStartReturns dials once without retry: Start must return only when the port accepts connections.
func requirePortIsBoundWhenStartReturns[T starter](t *testing.T, create func(port string) T) {
	for i := 0; i < 200; i++ {
		port, s := startOnFreePort(t, create)

		conn, err := net.Dial("tcp", "localhost:"+port)
		s.Stop()
		require.NoError(t, err, "iteration %d", i)
		_ = conn.Close()
	}
}

func requireStartReturnsErrorWhenPortIsInUse[T starter](t *testing.T, create func(port string) T) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer ln.Close()
	port := fmt.Sprintf("%v", ln.Addr().(*net.TCPAddr).Port)

	s := create(port)
	require.Error(t, s.Start())
	s.Stop() // must not panic after a failed start
}

// requireStopFreesPort waits for the port because a listener handed to Serve after Stop is closed asynchronously.
func requireStopFreesPort[T starter](t *testing.T, create func(port string) T) {
	port, s := startOnFreePort(t, create)
	s.Stop()

	require.Eventually(t, func() bool {
		ln, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return false
		}
		_ = ln.Close()
		return true
	}, 5*time.Second, 10*time.Millisecond)
}
