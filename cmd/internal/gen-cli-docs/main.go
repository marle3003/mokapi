package main

import (
	"fmt"
	"mokapi/pkg/cmd/mokapi"
)

func main() {
	cmd := mokapi.NewCmdMokapi()
	err := cmd.GenMarkdown("./docs/configuration/static")
	if err != nil {
		fmt.Println(err.Error())
	}
}
