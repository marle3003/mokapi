package kafka

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	luam "mokapi/lua"
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
	message := luam.MapTable(state.ToTable(4))
	partition := -1
	if lv, ok := state.Get(5).(lua.LNumber); ok {
		partition = int(lv)
	}

	k, m, err := kafka.producer(broker, topic, partition, key, message)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 3
	}

	state.Push(luar.New(state, k))
	state.Push(luar.New(state, m))

	return 2
}

func (kafka *Kafka) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"produce": kafka.Produce,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}
