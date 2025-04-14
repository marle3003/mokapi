package file

import (
	"encoding/json"
	"github.com/dop251/goja"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
}

func Enable(rt *goja.Runtime, host common.Host) {
	r := &Module{
		host: host,
		rt:   rt,
	}
	rt.Set("open", r.open)
}

func (o *Module) open(file string, args map[string]interface{}) (any, error) {
	f, err := o.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	switch args["as"] {
	case "binary":
		return f.Raw, nil
	case "resolved":
		return o.resolve(f)
	case "string":
		fallthrough
	default:
		return string(f.Raw), nil
	}
}

func (o *Module) resolve(f *dynamic.Config) (any, error) {
	b, err := json.Marshal(f.Data)
	if err != nil {
		return nil, err
	}
	var v any
	err = json.Unmarshal(b, &v)
	return v, err
}
