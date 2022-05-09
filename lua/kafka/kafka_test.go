package kafka

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"testing"
)

type client struct {
	produce func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error)
}

func TestModule_Produce(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, l *lua.LState, m *Module, c *client)
	}{
		{
			"cluster should be foo",
			func(t *testing.T, l *lua.LState, m *Module, c *client) {
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.Equal(t, "foo", cluster)
					return nil, nil, nil
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
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.Equal(t, "foo", topic)
					return nil, nil, nil
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
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.Equal(t, -1, partition)
					return nil, nil, nil
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
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.Equal(t, 10, partition)
					return nil, nil, nil
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
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.Equal(t, "foo", key)
					return nil, nil, nil
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
				c.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					require.IsType(t, &sortedmap.LinkedHashMap{}, value)
					require.Equal(t, "bar", value.(*sortedmap.LinkedHashMap).Get("foo"))
					return nil, nil, nil
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

func (c *client) Produce(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
	if c.produce != nil {
		return c.produce(cluster, topic, partition, key, value, headers)
	}
	return nil, nil, nil
}
