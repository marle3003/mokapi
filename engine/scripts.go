package engine

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/lua"
	"path/filepath"
)

type ScriptLoader interface {
	Load(file *dynamic.Config, host common.Host) (common.Script, error)
}

type ScriptLoaderFunc func(file *dynamic.Config, host common.Host) (common.Script, error)

type DefaultScriptLoader struct {
	config *static.Config
}

func (f ScriptLoaderFunc) Load(file *dynamic.Config, host common.Host) (common.Script, error) {
	return f(file, host)
}

func NewDefaultScriptLoader(config *static.Config) ScriptLoader {
	return &DefaultScriptLoader{config: config}
}

func (l *DefaultScriptLoader) Load(file *dynamic.Config, host common.Host) (common.Script, error) {
	s := file.Data.(string)
	filename := file.Info.Path()
	switch filepath.Ext(filename) {
	case ".js", ".cjs", ".mjs", ".ts":
		return js.New(file, host, l.config.Js)
	case ".lua":
		return lua.New(getScriptPath(file.Info.Url), s, host)
	default:
		return nil, fmt.Errorf("unsupported script %v", filename)
	}
}
