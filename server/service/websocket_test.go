package service_test

import (
	"fmt"
	"mokapi/server/service"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebsocketServer(t *testing.T) {
	t.Parallel()
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	b := service.NewWebsocketServer(fmt.Sprintf("%v", port), handler)
	b.Start()
	defer b.Stop()

	try.GetRequest(t, fmt.Sprintf("http://%s", addr), nil,
		try.HasStatusCode(200),
	)
	require.True(t, called, "handler should be called")
}
