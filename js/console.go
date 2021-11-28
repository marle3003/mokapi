package js

import (
	"github.com/dop251/goja"
	"mokapi/js/common"
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
	runtime.Set("console", common.Map(runtime, c))
}

func (c *console) Log(msg string) {
	c.logger.Info(msg)
}

func (c *console) Warn(msg string) {
	c.logger.Warn(msg)
}

func (c *console) Error(msg string) {
	c.logger.Error(msg)
}
