package main

import (
	"context"
	"fmt"
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
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
