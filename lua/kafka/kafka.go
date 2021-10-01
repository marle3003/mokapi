package kafka

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"mokapi/lua/utils"
	"strings"
	"time"
)

type WriteMessage func(broker, topic string, partition int, key, message interface{}) (interface{}, interface{}, error)

type Kafka struct {
	producer WriteMessage
}

func NewKafka(write WriteMessage) *Kafka {
	return &Kafka{
		producer: write,
	}
}

func (kafka *Kafka) Produce(state *lua.LState) int {
	broker := state.CheckString(1)
	topic := state.CheckString(2)
	key := state.ToString(3)
	message := utils.MapTable(state.ToTable(4))
	partition := -1
	if lv, ok := state.Get(5).(lua.LNumber); ok {
		partition = int(lv)
	}
	timeout := time.Second * 30
	if lv, ok := state.Get(6).(lua.LString); ok {
		if d, err := time.ParseDuration(string(lv)); err != nil {
			state.Push(lua.LNil)
			state.Push(lua.LNil)
			state.Push(lua.LString(err.Error()))
			return 3
		} else {
			timeout = d
		}
	}

	var err error
	var k, m interface{}
	for start := time.Now(); time.Since(start) < timeout; {
		if k, m, err = kafka.producer(broker, topic, partition, key, message); err == nil {
			state.Push(luar.New(state, k))
			state.Push(luar.New(state, m))
			return 2
		} else if !strings.HasPrefix(err.Error(), "no broker found at") {
			break
		}
	}

	state.Push(lua.LNil)
	state.Push(lua.LNil)
	state.Push(lua.LString(err.Error()))
	return 3
}

func (kafka *Kafka) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"produce": kafka.Produce,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}
