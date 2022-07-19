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

func (o *open) open(file string) (string, error) {
	_, s, err := o.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	return s, nil
}
