package file

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/engine/common"

	"github.com/dop251/goja"
)

type Module struct {
	host   common.Host
	rt     *goja.Runtime
	parent *dynamic.Config
}

func Enable(rt *goja.Runtime, host common.Host, parent *dynamic.Config) {
	r := &Module{
		host:   host,
		rt:     rt,
		parent: parent,
	}
	_ = rt.Set("open", r.open)
}

func (o *Module) open(file string, args map[string]interface{}) (any, error) {
	f, err := o.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	dynamic.AddRef(o.parent, f)
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
