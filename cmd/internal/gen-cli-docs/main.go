package main

import (
	"fmt"
	"mokapi/pkg/cmd/mokapi"
)

func main() {
	cmd := mokapi.NewCmdMokapi()
	err := cmd.GenMarkdown("./docs/cli")
	if err != nil {
		fmt.Println(err.Error())
	}
}
