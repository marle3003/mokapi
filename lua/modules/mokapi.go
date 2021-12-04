package modules

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/engine/common"
)

type Mokapi struct {
	host common.Host
}

func NewMokapi(host common.Host) *Mokapi {
	return &Mokapi{host: host}
}

func (m *Mokapi) Every(every string, do func(), args ...interface{}) (int, error) {
	times := -1
	if len(args) > 0 {
		times = int(args[0].(float64))
	}

	return m.host.Every(every, do, times)
}

func (m *Mokapi) On(event string, do func(args ...interface{}) bool, args ...interface{}) {
	tags := make(map[string]string)

	if len(args) > 0 {
		m := args[0].(map[interface{}]interface{})
		for f, v := range m {
			switch f.(string) {
			case "tags":
				for k, v := range v.(map[interface{}]interface{}) {
					tags[k.(string)] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	f := func(args ...interface{}) (bool, error) {
		var err error
		b := func() bool {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("error in script %v", r)
				}
			}()

			return do(args...)
		}()

		return b, err
	}

	m.host.On(event, f, tags)
}

func (m *Mokapi) Loader(state *lua.LState) int {
	state.Push(luar.New(state, m))
	return 1
}
