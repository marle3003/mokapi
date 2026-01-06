package main

import (
	"context"
	"fmt"
	"mokapi/pkg/cmd/mokapi"
)

func main() {
	cmd := mokapi.NewCmdMokapi(context.Background())
	err := cmd.GenMarkdown("./docs/cli")
	if err != nil {
		fmt.Println(err.Error())
	}
}
