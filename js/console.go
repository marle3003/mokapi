package js

import (
	"encoding/json"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"strings"
)

type console struct {
	runtime *goja.Runtime
	logger  common.Logger
}

func enableConsole(runtime *goja.Runtime, logger common.Logger) {
	c := &console{
		runtime: runtime,
		logger:  logger,
	}
	runtime.Set("console", mapToJSValue(runtime, c))
}

func (c *console) Log(args ...goja.Value) {
	c.log(logrus.InfoLevel, args...)
}

func (c *console) Warn(args ...goja.Value) {
	c.log(logrus.WarnLevel, args...)
}

func (c *console) Error(args ...goja.Value) {
	c.log(logrus.ErrorLevel, args...)
}

func (c *console) Debug(args ...goja.Value) {
	c.log(logrus.DebugLevel, args...)
}

func (c *console) log(level logrus.Level, args ...goja.Value) {
	var sb strings.Builder
	for i, arg := range args {
		if i > 1 {
			sb.WriteString(" ")
		}
		sb.WriteString(c.toString(arg))
	}
	msg := sb.String()

	switch level {
	case logrus.InfoLevel:
		c.logger.Info(msg)
	case logrus.WarnLevel:
		c.logger.Warn(msg)
	case logrus.ErrorLevel:
		c.logger.Error(msg)
	case logrus.DebugLevel:
		c.logger.Debug(msg)
	}
}

func (c *console) toString(v goja.Value) string {
	m, ok := v.(json.Marshaler)
	if !ok {
		return v.String()
	}

	b, err := json.Marshal(m)
	if err != nil {
		return v.String()
	}
	return string(b)
}
