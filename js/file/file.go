package file

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
)

type Module struct {
	host common.Host
}

func Enable(rt *goja.Runtime, host common.Host) {
	r := &Module{
		host: host,
	}
	rt.Set("open", r.open)
}

func (o *Module) open(file string, args map[string]interface{}) (interface{}, error) {
	f, err := o.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	switch args["as"] {
	case "binary":
		return f.Raw, nil
	case "string":
		fallthrough
	default:
		return string(f.Raw), nil
	}
}
