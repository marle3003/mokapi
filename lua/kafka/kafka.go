package kafka

import (
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/engine/common"
	"mokapi/lua/convert"
	"strings"
	"time"
)

type Module struct {
	client common.KafkaClient
}

type produceOptions struct {
	Cluster   string
	Topic     string
	Partition int
	Key       interface{}
	Value     interface{}
	Headers   map[string]interface{}
	Timeout   int
}

func New(c common.KafkaClient) *Module {
	return &Module{
		client: c,
	}
}

func (m *Module) Produce(state *lua.LState) int {
	args := &common.KafkaProduceArgs{Timeout: 30, Partition: -1}
	if lArg := state.Get(1); lArg != lua.LNil {
		if err := convert.FromLua(lArg, &args); err != nil {
			log.Error(err)
		}
	}

	var err error
	timeout := time.Duration(args.Timeout) * time.Second
	for start := time.Now(); time.Since(start) < timeout; {
		if result, err := m.client.Produce(args); err == nil {
			state.Push(luar.New(state, result))
			return 1
		} else if !strings.HasPrefix(err.Error(), "no broker found at") {
			break
		}
	}

	state.Push(lua.LNil)
	state.Push(lua.LString(err.Error()))
	return 2
}

func (m *Module) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"produce": m.Produce,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}
