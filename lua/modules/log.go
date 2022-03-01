package modules

import (
	lua "github.com/yuin/gopher-lua"
	"mokapi/engine/common"
)

type Log struct {
	host common.Host
}

func NewLog(host common.Host) *Log {
	l := &Log{host: host}
	return l
}

func (l *Log) Info(s string) {
	l.host.Info(s)
}

func (l *Log) Warn(s string) {
	l.host.Warn(s)
}

func (l *Log) Error(s string) {
	l.host.Error(s)
}

func (l *Log) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"info": func(L *lua.LState) int {
			s := L.ToString(1)
			l.Info(s)
			return 0
		},
		"warn": func(L *lua.LState) int {
			s := L.ToString(1)
			l.Warn(s)
			return 0
		},
		"error": func(L *lua.LState) int {
			s := L.ToString(1)
			l.Error(s)
			return 0
		},
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}
