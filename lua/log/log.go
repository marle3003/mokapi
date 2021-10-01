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

func (l *Log) Info(s string) {
	log.Info(s)
	if l.Handler != nil && l.Level >= InfoLevel {
		l.Handler(fmt.Sprintf("info: %v", s))
	}
}

func (l *Log) Warn(s string) {
	log.Warn(s)
	if l.Handler != nil && l.Level >= WarnLevel {
		l.Handler(fmt.Sprintf("warn: %v", s))
	}
}

func (l *Log) Error(s string) {
	log.Error(s)
	l.Handler(fmt.Sprintf("error: %v", s))
}

func (l *Log) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"debug": func(L *lua.LState) int {
			s := L.ToString(1)
			l.Debug(s)
			return 0
		},
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
