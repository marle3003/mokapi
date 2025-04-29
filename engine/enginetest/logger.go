package enginetest

type Logger struct {
	InfoFunc           func(args ...interface{})
	WarnFunc           func(args ...interface{})
	ErrorFunc          func(args ...interface{})
	DebugFunc          func(args ...interface{})
	IsLevelEnabledFunc func(level string) bool
}

func (l *Logger) Info(args ...interface{}) {
	if l.InfoFunc == nil {
		return
	}
	l.InfoFunc(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	if l.WarnFunc == nil {
		return
	}
	l.WarnFunc(args...)
}

func (l *Logger) Error(args ...interface{}) {
	if l.ErrorFunc == nil {
		return
	}
	l.ErrorFunc(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	if l.DebugFunc == nil {
		return
	}
	l.DebugFunc(args...)
}

func (l *Logger) IsLevelEnabled(level string) bool {
	if l.IsLevelEnabledFunc == nil {
		return true
	}
	return l.IsLevelEnabledFunc(level)
}
