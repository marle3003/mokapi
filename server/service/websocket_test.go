package service_test

import (
	"fmt"
	"io"
	"mokapi/server/service"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestWebsocketServer(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Parallel()
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	b := service.NewWebsocketServer(fmt.Sprintf("%v", port), handler)
	require.NoError(t, b.Start())
	defer b.Stop()

	try.GetRequest(t, fmt.Sprintf("http://%s", addr), nil,
		try.HasStatusCode(200),
	)
	require.True(t, called, "handler should be called")
}

func newWebsocketServer(port string) *service.WebsocketServer {
	return service.NewWebsocketServer(port, http.NotFoundHandler())
}

func TestWebsocketServer_Start_PortIsBoundWhenStartReturns(t *testing.T) {
	requirePortIsBoundWhenStartReturns(t, newWebsocketServer)
}

func TestWebsocketServer_Start_ReturnsErrorWhenPortIsInUse(t *testing.T) {
	requireStartReturnsErrorWhenPortIsInUse(t, newWebsocketServer)
}

func TestWebsocketServer_Stop_FreesPort(t *testing.T) {
	requireStopFreesPort(t, newWebsocketServer)
}

func TestWebsocketServer_Stop_LogsNoError(t *testing.T) {
	hook := logrustest.NewGlobal()
	defer hook.Reset()

	s := newWebsocketServer(fmt.Sprintf("%v", try.GetFreePort()))
	require.NoError(t, s.Start())
	s.Stop()

	// Serve runs in a goroutine and logs after Stop returned
	time.Sleep(100 * time.Millisecond)
	for _, e := range hook.AllEntries() {
		require.NotEqual(t, logrus.ErrorLevel, e.Level, e.Message)
	}
}
