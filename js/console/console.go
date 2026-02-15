package console

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
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
	if len(args) == 0 {
		return
	}

	var out []string

	if format, ok := args[0].Export().(string); ok && strings.Contains(format, "%") {
		specs := countFormatSpecifiers(format)

		var exported []any
		for _, arg := range args[1:] {
			exported = append(exported, export(arg))
		}

		if specs > 0 && len(exported) >= specs {
			if s, ok := formatString(format, exported[:specs]...); ok {
				out = append(out, s)

				// Append remaining args (browser behavior)
				for _, arg := range exported[specs:] {
					out = append(out, fmt.Sprintf("%v", arg))
				}

				goto LOG
			}
		}
	}

	// Fallback: no formatting or formatting failed
	for _, arg := range args {
		out = append(out, fmt.Sprintf("%v", export(arg)))
	}

LOG:
	msg := strings.Join(out, " ")

	switch level {
	case logrus.WarnLevel:
		c.logger.Warn(msg)
	case logrus.ErrorLevel:
		c.logger.Error(msg)
	case logrus.DebugLevel:
		c.logger.Debug(msg)
	default:
		c.logger.Info(msg)
	}
}

func export(v goja.Value) any {
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

func countFormatSpecifiers(format string) int {
	count := 0
	escaped := false

	for i := 0; i < len(format); i++ {
		if escaped {
			escaped = false
			continue
		}
		if format[i] == '%' {
			if i+1 < len(format) && format[i+1] == '%' {
				escaped = true
				continue
			}
			count++
		}
	}
	return count
}
