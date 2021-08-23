package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

type level uint32

const (
	ErrorLevel level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

type logHandler func(s string)

type Log struct {
	Level   level
	Handler logHandler
}

func NewLog() *Log {
	l := &Log{}
	l.setLevel(log.GetLevel())
	return l
}

func (l *Log) Debug(s string) {
	log.Debug(s)
	if l.Handler != nil && l.Level >= DebugLevel {
		l.Handler(fmt.Sprintf("debug: %v", s))
	}
}

func (l *Log) Loader(state *lua.LState) int {
	debug := func(L *lua.LState) int {
		s := L.ToString(1)
		l.Debug(s)
		return 0
	}

	exports := map[string]lua.LGFunction{
		"debug": debug,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

func (l *Log) setLevel(level log.Level) {
	switch level {
	case log.ErrorLevel:
		l.Level = ErrorLevel
	case log.WarnLevel:
		l.Level = WarnLevel
	case log.InfoLevel:
		l.Level = InfoLevel
	case log.DebugLevel:
		l.Level = DebugLevel
	default:
		l.Level = ErrorLevel
	}
}
