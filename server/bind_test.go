package server_test

import (
	"fmt"
	"mokapi/try"
	"net"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

// requireBoundWhenUpdateReturns dials once without retry or sleep: the manager must have bound the port
// when UpdateConfig returns.
func requireBoundWhenUpdateReturns(t *testing.T, update func(port int)) {
	port := try.GetFreePort()
	update(port)

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err)
	_ = conn.Close()
}

// requireBindErrorIsLoggedAndServerStaysRegistered occupies the port, expects the bind error in the log
// and checks that a second update after the port is free does not add the server again.
// A manager that forgot the server would log expectedAddingLog a second time. The log is checked
// instead of dialing the freed port: on macOS a foreign process may listen on a specific address of it.
func requireBindErrorIsLoggedAndServerStaysRegistered(t *testing.T, update func(port int), expectedLog, expectedAddingLog func(port int) string) {
	hook := test.NewGlobal()
	defer hook.Reset()

	occupied, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer occupied.Close()
	port := occupied.Addr().(*net.TCPAddr).Port

	update(port)

	require.Contains(t, logMessages(hook), expectedLog(port))

	require.Contains(t, logMessages(hook), expectedAddingLog(port))

	_ = occupied.Close()
	hook.Reset()
	update(port)

	require.NotContains(t, logMessages(hook), expectedAddingLog(port))
}

func logMessages(hook *test.Hook) []string {
	var messages []string
	for _, e := range hook.AllEntries() {
		messages = append(messages, e.Message)
	}
	return messages
}
