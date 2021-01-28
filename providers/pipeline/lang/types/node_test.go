package types

import "testing"

func TestNode_GetField(t *testing.T) {
	type test func(*testing.T, Object)
	data := []struct {
		node   *Node
		field  string
		assert test
	}{
		{func() *Node {
			n := NewNode("foo")
			n.children = NewArray()
			c := NewNode("bar")
			c.content = "test"
			n.Add(c)
			return n
		}(),
			"bar",
			func(t *testing.T, obj Object) {
				if obj.String() != "<bar>test</bar>" {
					t.Errorf("got %v, expected %v", obj.String(), "test")
				}
			},
		},
	}

	for _, d := range data {
		obj, err := d.node.GetField(d.field)
		if err != nil {
			t.Errorf("Node Getfield(%q):%v", d.field, err)
		}
		d.assert(t, obj)
	}
}
