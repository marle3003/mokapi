package lua

import (
	log "github.com/sirupsen/logrus"
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
	src   string
}

func New(filename, src string, host common.Host) (*Script, error) {
	script := &Script{Key: filename, src: src}
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
	script.state.PreloadModule("http", http.New().Loader)

	return script, nil
}

func (s *Script) Run() error {
	err := s.state.DoString(s.src)
	if err != nil {
		log.Errorf("script error %q: %v", s.Key, err)
	}
	return nil
}

func (s *Script) Close() {
	if s.state != nil && !s.state.IsClosed() {
		s.state.Close()
	}
}
