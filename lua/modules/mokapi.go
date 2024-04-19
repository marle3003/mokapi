package modules

import (
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"mokapi/engine/common"
	"mokapi/lua/convert"
	"reflect"
)

type Mokapi struct {
	host common.Host
}

type everyArgs struct {
	Times int
	Tags  map[string]string
}

type onArgs struct {
	Tags map[string]string
}

func NewMokapi(host common.Host) *Mokapi {
	return &Mokapi{host: host}
}

func (m *Mokapi) every(l *lua.LState) int {
	every := l.ToString(1)

	s := l.ToFunction(2)
	fn := func() {
		co, cocancel := l.NewThread()
		defer func() {
			if cocancel != nil {
				cocancel()
			}
		}()
		_, err, values := l.Resume(co, s)
		_ = values

		if err != nil {
			panic(err)
		}
	}

	args := &everyArgs{}
	if lArg := l.Get(3); lArg != lua.LNil {
		if err := convert.FromLua(lArg, &args); err != nil {
			log.Error(err)
		}
	}

	opt := common.JobOptions{
		Times:                 args.Times,
		Tags:                  args.Tags,
		SkipImmediateFirstRun: false,
	}
	id, err := m.host.Every(every, fn, opt)

	if err != nil {
		l.Push(lua.LNumber(id))
		l.Push(lua.LString(err.Error()))
		return 2
	}

	l.Push(lua.LNumber(id))
	return 1
}

func (m *Mokapi) on(l *lua.LState) int {
	evt := l.CheckString(1)

	s := l.ToFunction(2)
	fn := func(args ...interface{}) (bool, error) {
		values := make([]lua.LValue, 0, len(args))
		for _, arg := range args {
			i, err := convert.ToLua(l, arg)
			if err != nil {
				return false, err
			}
			values = append(values, i)
		}

		co, cocancel := l.NewThread()
		defer func() {
			if cocancel != nil {
				cocancel()
			}
		}()
		_, err, retValues := l.Resume(co, s, values...)

		if err != nil {
			return false, err
		}

		if retValues[0] == lua.LTrue {
			for i, val := range values {
				// update only pointers
				if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
					err = convert.FromLua(val, args[i])
					if err != nil {
						return false, err
					}
				}

			}
		}

		return retValues[0] == lua.LTrue, nil
	}

	args := &onArgs{}
	if lArg := l.Get(3); lArg != lua.LNil {
		if err := convert.FromLua(lArg, &args); err != nil {
			log.Error(err)
		}
	}

	m.host.On(evt, fn, args.Tags)

	return 0
}

func (m *Mokapi) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"every": m.every,
		"on":    m.on,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}
