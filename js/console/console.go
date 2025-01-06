package console

import (
	"encoding/json"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"strings"
)

type Module struct {
	rt     *goja.Runtime
	logger common.Logger
}

func Enable(vm *goja.Runtime) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	f := &Module{
		rt:     vm,
		logger: host,
	}
	obj := vm.NewObject()
	obj.Set("log", f.Log)
	obj.Set("warn", f.Warn)
	obj.Set("error", f.Error)
	obj.Set("debug", f.Debug)
	vm.Set("console", obj)
}

func (c *Module) Log(args ...goja.Value) {
	c.log(logrus.InfoLevel, args...)
}

func (c *Module) Warn(args ...goja.Value) {
	c.log(logrus.WarnLevel, args...)
}

func (c *Module) Error(args ...goja.Value) {
	c.log(logrus.ErrorLevel, args...)
}

func (c *Module) Debug(args ...goja.Value) {
	c.log(logrus.DebugLevel, args...)
}

func (c *Module) log(level logrus.Level, args ...goja.Value) {
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

func (c *Module) toString(v goja.Value) string {
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
