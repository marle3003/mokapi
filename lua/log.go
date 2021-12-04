package lua

import (
	"mokapi/engine/common"
)

type Log struct {
	host common.Host
}

func newLog(host common.Host) *Log {
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
