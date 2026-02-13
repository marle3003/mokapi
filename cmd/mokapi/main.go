package main

import (
	"context"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cmd := mokapi.NewCmdMokapi()
	err := cmd.ExecuteWithContext(ctx)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}
}
