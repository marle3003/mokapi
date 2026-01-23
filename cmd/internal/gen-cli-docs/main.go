package main

import (
	"mokapi/pkg/cmd/mokapi"
	"os"
)

func main() {
	err := mokapi.NewGenCliDocCmd().Execute()
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
	}
}
