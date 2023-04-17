package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
)

type open struct {
	host common.Host
}

func enableOpen(rt *goja.Runtime, host common.Host) {
	r := &open{
		host: host,
	}
	rt.Set("open", r.open)
}

func (o *open) open(file string, args map[string]interface{}) (interface{}, error) {
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
