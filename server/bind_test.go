package server_test

import (
	"fmt"
	"mokapi/try"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
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
// and checks that the server is not created again by a second update after the port is free.
// A manager that forgot the server would bind the freed port then.
// Pass checkRegistered=false for servers that can't be found again by the manager (see MailManager keys).
func requireBindErrorIsLoggedAndServerStaysRegistered(t *testing.T, checkRegistered bool, update func(port int), expectedLog func(port int) string) {
	hook := test.NewGlobal()
	defer hook.Reset()

	occupied, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer occupied.Close()
	port := occupied.Addr().(*net.TCPAddr).Port

	update(port)

	require.Contains(t, errorMessages(hook), expectedLog(port))

	if !checkRegistered {
		return
	}
	_ = occupied.Close()
	update(port)
	_, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	require.Error(t, err)
}

func errorMessages(hook *test.Hook) []string {
	var messages []string
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.ErrorLevel {
			messages = append(messages, e.Message)
		}
	}
	return messages
}
