package faker

import (
	"github.com/dop251/goja"
	"mokapi/schema/json/generator"
)

type Node struct {
	origNode   *generator.Node
	children   goja.Value
	attributes goja.Value
	dependsOn  goja.Value

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
		if n.children == nil {
			n.children = newJsArray[*generator.Node](&children{n: n}, n.m.vm, func(v goja.Value) *generator.Node {
				return convertToNode(v, n.m)
			})
		}
		return n.children
	case "attributes":
		if n.attributes == nil {
			n.attributes = newJsArray[string](&attributes{n: n}, n.m.vm, func(v goja.Value) string {
				return v.String()
			})
		}
		return n.children
	case "dependsOn":
		if n.dependsOn == nil {
			n.dependsOn = newJsArray[string](&dependsOn{n: n}, n.m.vm, func(v goja.Value) string {
				return v.String()
			})
		}
		return n.dependsOn
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
	case "attributes":
		values, ok := val.Export().([]string)
		if ok {
			old := n.origNode.Attributes
			n.origNode.Attributes = values
			n.restore = append(n.restore, func() {
				n.origNode.Attributes = old
			})
		}
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
	result, restore := Set(c.n.origNode.Children, i, val)
	c.n.origNode.Children = result
	c.n.restore = append(c.n.restore, func() {
		c.n.origNode.Children = restore()
	})
}

func (c *children) Splice(start int, deleteCount int, items []*generator.Node) {
	result, restore := splice(c.n.origNode.Children, start, deleteCount, items)
	c.n.origNode.Children = result
	c.n.restore = append(c.n.restore, func() {
		c.n.origNode.Children = restore()
	})
}

type attributes struct {
	n *Node
}

func (a *attributes) Len() int {
	return len(a.n.origNode.Children)
}

func (a *attributes) Get(i int) string {
	return a.n.origNode.Attributes[i]
}

func (a *attributes) Set(i int, val string) {
	result, restore := Set(a.n.origNode.Attributes, i, val)
	a.n.origNode.Attributes = result
	a.n.restore = append(a.n.restore, func() {
		a.n.origNode.Attributes = restore()
	})
}

func (a *attributes) Splice(start int, deleteCount int, items []string) {
	result, restore := splice(a.n.origNode.Attributes, start, deleteCount, items)
	a.n.origNode.Attributes = result
	a.n.restore = append(a.n.restore, func() {
		a.n.origNode.Attributes = restore()
	})
}

type dependsOn struct {
	n *Node
}

func (d *dependsOn) Len() int {
	return len(d.n.origNode.DependsOn)
}

func (d *dependsOn) Get(i int) string {
	return d.n.origNode.DependsOn[i]
}

func (d *dependsOn) Set(i int, val string) {
	result, restore := Set(d.n.origNode.DependsOn, i, val)
	d.n.origNode.DependsOn = result
	d.n.restore = append(d.n.restore, func() {
		d.n.origNode.DependsOn = restore()
	})
}

func (d *dependsOn) Splice(start int, deleteCount int, items []string) {
	result, restore := splice(d.n.origNode.DependsOn, start, deleteCount, items)
	d.n.origNode.DependsOn = result
	d.n.restore = append(d.n.restore, func() {
		d.n.origNode.DependsOn = restore()
	})
}
