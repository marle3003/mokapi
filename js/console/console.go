package console

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"regexp"
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
	format := ""
	if len(args) > 0 {
		if s, ok := args[0].Export().(string); ok && strings.Contains(s, "%") {
			format = s
		}
	}

	var logArgs []any
	for _, arg := range args {
		logArgs = append(logArgs, c.format(arg))
	}

	if format != "" {
		if s, ok := formatString(format, logArgs[1:]...); ok {
			logArgs = []any{s}
		}
	}

	switch level {
	case logrus.WarnLevel:
		c.logger.Warn(logArgs...)
	case logrus.ErrorLevel:
		c.logger.Error(logArgs...)
	case logrus.DebugLevel:
		c.logger.Debug(logArgs...)
	default:
		c.logger.Info(logArgs...)
	}
}

func (c *Module) format(v goja.Value) any {
	m, ok := v.(json.Marshaler)
	if !ok {
		return v.Export()
	}

	b, err := json.Marshal(m)
	if err != nil {
		return v.Export()
	}
	return string(b)
}

var fmtErrorRegex = regexp.MustCompile(`%!(?:\([A-Z]+[^)]*\)|[a-zA-Z]\(MISSING\))`)

func formatString(format string, args ...any) (string, bool) {
	out := fmt.Sprintf(format, args...)

	// Detect Go's error output patterns
	if fmtErrorRegex.MatchString(out) {
		return "", false
	}

	return out, true
}
