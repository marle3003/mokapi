package lua

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/engine/common"
	"mokapi/lua/http"
	"mokapi/lua/kafka"
	"mokapi/lua/modules"
	"path/filepath"
	"runtime/debug"
	"strings"
)

type Script struct {
	Key   string
	state *lua.LState
	src   string
}

func New(filename, src string, host common.Host) (*Script, error) {
	// add directory of script to package path. Lua uses this for searching modules
	sep := string(filepath.Separator)
	pkg := fmt.Sprintf("package.path = package.path .. \";%v%v\" .. [[%v?.lua]]\n", filepath.Dir(filename), sep, sep)
	pkg = strings.ReplaceAll(pkg, "\\", "\\\\")
	src = pkg + src

	script := &Script{Key: filename, src: src}
	script.state = lua.NewState(lua.Options{IncludeGoStackTrace: true})
	script.state.SetGlobal("script_path", lua.LString(filename))
	script.state.SetGlobal("script_dir", lua.LString(filepath.Dir(filename)))
	script.state.SetGlobal("dump", luar.New(script.state, Dump))
	script.state.SetGlobal("sleep", luar.New(script.state, sleep))
	script.state.SetGlobal("open", luar.New(script.state, newFile(host).open))
	script.state.PreloadModule("log", modules.NewLog(host).Loader)
	script.state.PreloadModule("mokapi", modules.NewMokapi(host).Loader)
	script.state.PreloadModule("yaml", modules.YamlLoader)
	script.state.PreloadModule("kafka", kafka.New(host.KafkaClient()).Loader)
	script.state.PreloadModule("mustache", modules.MustacheLoader)
	script.state.PreloadModule("http", http.New().Loader)

	return script, nil
}

func (s *Script) Run() error {
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("lua script error: %v", string(debug.Stack()))
			log.Errorf("lua script error: %v", r)
		}
	}()

	err := s.state.DoString(s.src)
	if err != nil {
		return fmt.Errorf("syntax error %q: %v", s.Key, err)
	}
	return nil
}

func (s *Script) Close() {
	if s.state != nil && !s.state.IsClosed() {
		s.state.Close()
	}
}
