package mokapi

import (
	"mokapi/engine/common"

	"github.com/dop251/goja"
)

type App struct {
	api string
	m   *Module
}

func (a *App) Http() goja.Value {
	h := &Http{filter: common.HttpFilter{Api: a.api}, m: a.m}
	do := a.m.vm.NewDynamicObject(&HttpObject{
		http: h,
		vm:   h.m.vm,
	})
	return do
}

func (a *App) Api(name string) *App {
	return &App{api: name, m: a.m}
}
