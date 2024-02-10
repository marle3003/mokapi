package webui

import (
	"embed"
)

//go:embed all:dist
var App embed.FS
