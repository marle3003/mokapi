package kafka

import (
	"fmt"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/engine/common"
	"mokapi/sortedmap"
	"testing"
)

type client struct {
	produce func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error)
}

func TestModule_Produce(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, l *lua.LState, m *Module, c *client)
	}{
		{
			"cluster should be foo",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.Equal(t, "foo", args.Cluster)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce({cluster="foo"})`,
				)
				require.NoError(t, err)
			},
		},
		{
			"topic should be foo",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.Equal(t, "foo", args.Topic)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce({topic="foo"})`,
				)
				require.NoError(t, err)
			},
		},
		{
			"partition should be -1",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.Equal(t, -1, args.Messages[0].Partition)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce()`,
				)
				require.NoError(t, err)
			},
		},
		{
			"partition should be 10",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.Equal(t, 10, args.Messages[0].Partition)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce({partition=10})`,
				)
				require.NoError(t, err)
			},
		},
		{
			"key should be foo",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.Equal(t, "foo", args.Messages[0].Key)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce({key="foo"})`,
				)
				require.NoError(t, err)
			},
		},
		{
			"value",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					require.IsType(t, &sortedmap.LinkedHashMap[string, interface{}]{}, args.Messages[0].Data)
					foo, _ := args.Messages[0].Data.(*sortedmap.LinkedHashMap[string, interface{}]).Get("foo")
					require.Equal(t, "bar", foo)
					return nil, nil
				}
				err := l.DoString(`
					kafka = require("kafka")
					kafka.produce({value={foo="bar"}})`,
				)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &client{}
			m := New(c)
			l := lua.NewState()
			l.PreloadModule("kafka", m.Loader)
			tc.f(t, l, m, c)
		})
	}
}

func (c *client) Produce(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
	if c.produce != nil {
		return c.produce(args)
	}
	return nil, fmt.Errorf("function not defined")
}
