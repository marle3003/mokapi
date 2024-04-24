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
	*common.KafkaProduceArgs
	Partition int
	Key       interface{}
	Value     interface{}
	Headers   map[string]interface{}
	Timeout   int
}

type kafkaResult struct {
	*common.KafkaProduceResult

	Key       interface{}
	Value     interface{}
	Offset    int64
	Headers   map[string]string
	Partition int
}

func New(c common.KafkaClient) *Module {
	return &Module{
		client: c,
	}
}

func (m *Module) Produce(state *lua.LState) int {
	opts := &produceOptions{Timeout: 30, Partition: -1, KafkaProduceArgs: &common.KafkaProduceArgs{}}
	if lArg := state.Get(1); lArg != lua.LNil {
		if err := convert.FromLua(lArg, &opts); err != nil {
			log.Error(err)
		}
	}
	args := &common.KafkaProduceArgs{Cluster: opts.Cluster, Topic: opts.Topic, Timeout: opts.Timeout}
	args.Records = append(args.Records, common.KafkaRecord{
		Key:       opts.Key,
		Data:      opts.Value,
		Headers:   opts.Headers,
		Partition: opts.Partition,
	})

	var err error
	timeout := time.Duration(args.Timeout) * time.Second
	for start := time.Now(); time.Since(start) < timeout; {
		if result, err := m.client.Produce(args); err == nil {
			r := &kafkaResult{KafkaProduceResult: result}
			if result != nil && len(result.Records) == 1 {
				r.Key = result.Records[0].Key
				r.Value = result.Records[0].Value
				r.Headers = result.Records[0].Headers
				r.Partition = result.Records[0].Partition
			}
			state.Push(luar.New(state, r))
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
