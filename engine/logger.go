package engine

import log "github.com/sirupsen/logrus"

type logger struct {
	logger *log.Logger
}

func newLogger(l *log.Logger) *logger {
	return &logger{logger: l}
}

func (l *logger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *logger) IsLevelEnabled(level string) bool {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return false
	}
	return l.logger.IsLevelEnabled(lvl)
}
