package lua

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/engine/common"
	"mokapi/lua/http"
	"mokapi/lua/modules"
	"path/filepath"
)

type Script struct {
	Key   string
	state *lua.LState
}

func New(filename, src string, host common.Host) (*Script, error) {
	script := &Script{Key: filename}
	script.state = lua.NewState(lua.Options{IncludeGoStackTrace: true})
	script.state.SetGlobal("script_path", lua.LString(filename))
	script.state.SetGlobal("script_dir", lua.LString(filepath.Dir(filename)))
	script.state.SetGlobal("dump", luar.New(script.state, Dump))
	script.state.SetGlobal("sleep", luar.New(script.state, sleep))
	script.state.SetGlobal("open", luar.New(script.state, newFile(host).open))
	script.state.SetGlobal("log", luar.New(script.state, newLog(host)))
	script.state.PreloadModule("mokapi", modules.NewMokapi(host).Loader)
	script.state.PreloadModule("yaml", modules.YamlLoader)
	//l.state.PreloadModule("kafka", kafka.Loader)
	script.state.PreloadModule("mustache", modules.MustacheLoader)
	script.state.PreloadModule("http", http.Loader)

	err := script.state.DoString(src)
	if err != nil {
		return nil, fmt.Errorf("script error %q: %v", filename, err)
	}

	return script, nil
}

func (s *Script) Run() error {
	return nil
}

func (s *Script) Close() {
	if s.state != nil && !s.state.IsClosed() {
		s.state.Close()
	}
}
