package faker

import (
	"github.com/dop251/goja"
	"mokapi/schema/json/generator"
)

type Node struct {
	origNode *generator.Node
	children goja.Value

	m       *Module
	test    func(r *generator.Request) bool
	fake    func(r *generator.Request) (interface{}, error)
	restore []func()
}

func NewNode(m *Module, node *generator.Node) goja.Value {
	n := &Node{
		origNode: node,
		m:        m,
	}

	n.children = newJsArray[*generator.Node](&children{n: n}, n.m.vm, func(v goja.Value) *generator.Node {
		return convertToNode(v, n.m)
	})

	m.host.AddCleanupFunc(func() {
		n.Restore()
	})

	return m.vm.NewDynamicObject(n)
}

func (n *Node) Get(key string) goja.Value {
	switch key {
	case "name":
		return n.m.vm.ToValue(n.origNode.Name)
	case "children":
		return n.children
	case "restore":
		return n.m.vm.ToValue(func(goja.FunctionCall) goja.Value {
			n.Restore()
			return goja.Undefined()
		})
	}
	return goja.Undefined()
}

func (n *Node) Set(key string, val goja.Value) bool {
	switch key {
	case "name":
		name := val.String()
		old := n.origNode.Name
		n.origNode.Name = name
		n.restore = append(n.restore, func() {
			if n.origNode.Name == name {
				n.origNode.Name = old
			}
		})
		return true
	}
	return false
}

func (n *Node) Delete(key string) bool {
	return true
}

func (n *Node) Has(key string) bool {
	switch key {
	case "name":
		return true
	default:
		return false
	}
}

func (n *Node) Keys() []string {
	return []string{
		"name",
	}
}

func (n *Node) Restore() {
	for _, f := range n.restore {
		f()
	}
}

type children struct {
	n *Node
}

func (c *children) Len() int {
	return len(c.n.origNode.Children)
}

func (c *children) Get(i int) *generator.Node {
	return c.n.origNode.Children[i]
}

func (c *children) Set(i int, val *generator.Node) {
	if i < len(c.n.origNode.Children) {
		old := c.n.origNode.Children[i]
		c.n.origNode.Children[i] = val
		c.n.restore = append(c.n.restore, func() { c.n.origNode.Children[i] = old })
	} else {
		n := len(c.n.origNode.Children)
		// Fill with nil nodes if needed
		for len(c.n.origNode.Children) < i {
			c.n.origNode.Children = append(c.n.origNode.Children, nil)
		}
		c.n.origNode.Children = append(c.n.origNode.Children, val)
		c.n.restore = append(c.n.restore, func() {
			c.n.origNode.Children = c.n.origNode.Children[:n]
		})
	}
}

func (c *children) Splice(start int, deleteCount int, items []*generator.Node) {
	result, restore := splice(c.n.origNode.Children, start, deleteCount, items)
	c.n.origNode.Children = result
	c.n.restore = append(c.n.restore, func() {
		c.n.origNode.Children = restore()
	})
}

func splice[T any](array []T, start int, deleteCount int, items []T) ([]T, func() []T) {
	if start < 0 {
		return array, nil
	}

	end := start + deleteCount
	if end > len(array) {
		end = len(array)
	}

	var toAdd []T
	for _, item := range items {
		toAdd = append(toAdd, item)
	}

	removed := array[start:end]

	result := make([]T, 0, len(array)+len(toAdd)-deleteCount)
	result = append(result, array[:start]...)
	result = append(result, toAdd...)
	result = append(result, array[end:]...)

	restore := func() []T {
		restore := make([]T, start)
		copy(restore, result[:start])

		restore = append(restore, removed...)

		added := len(toAdd)
		restore = append(restore, result[start+added:]...)
		return restore
	}

	return result, restore
}
