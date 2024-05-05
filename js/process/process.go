package process

import (
	"github.com/dop251/goja"
	"os"
	"strings"
)

type Module struct {
	env map[string]string
}

func Enable(runtime *goja.Runtime) {
	p := &Module{env: map[string]string{}}
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		p.env[kv[0]] = kv[1]
	}
	o := runtime.NewObject()
	o.Set("env", p.env)
	runtime.Set("process", o)
}
