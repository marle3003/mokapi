package main

import (
	"context"
	"mokapi/pkg/cmd/mokapi"
	"os"
)

func main() {
	cmd := mokapi.NewCmdMokapi()
	err := cmd.ExecuteWithContext(context.Background())
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}
}
