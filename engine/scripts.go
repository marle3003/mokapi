package engine

import (
	"errors"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/js"
	"mokapi/lua"
	"path/filepath"
)

var UnsupportedError = errors.New("unsupported script")

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
	if file.Info.Provider == "git" {
		gitFile := file.Info.Url.Query()["file"]
		if len(gitFile) > 0 {
			filename = gitFile[0]
		}
	}
	ext := filepath.Ext(filename)
	switch ext {
	case ".js", ".cjs", ".mjs", ".ts":
		return js.New(file, host)
	case ".lua":
		return lua.New(getScriptPath(file.Info.Url), s, host)
	default:
		return nil, UnsupportedError
	}
}
