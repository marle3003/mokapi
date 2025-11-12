package faker

import (
	"fmt"
	"mokapi/js/util"
	"mokapi/schema/json/generator"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
)

type Node struct {
	origNode   *generator.Node
	children   goja.Value
	attributes goja.Value
	dependsOn  goja.Value

	m       *Module
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
		return n.attributes
	case "weight":
		return n.m.vm.ToValue(n.origNode.Weight)
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
	case "fake":
		return n.m.vm.ToValue(n.origNode.Fake)
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
		if val.ExportType().Kind() != reflect.Slice {
			s := fmt.Sprintf("unexpected type for 'children': got %s, expected Array", util.JsType(val))
			panic(n.m.vm.ToValue(s))
		}
		arr := val.ToObject(n.m.vm)
		length := int(arr.Get("length").ToInteger())
		old := n.origNode.Attributes
		for i := 0; i < length; i++ {
			item := arr.Get(strconv.Itoa(i))
			if attr, ok := item.Export().(string); !ok {
				s := fmt.Sprintf("unexpected type for 'attributes[%d]': got %s, expected Array", i, util.JsType(val))
				panic(n.m.vm.ToValue(s))
			} else {
				n.origNode.Attributes = append(n.origNode.Attributes, attr)
			}
		}
		n.restore = append(n.restore, func() {
			n.origNode.Attributes = old
		})
		return true
	case "weight":
		t := val.ExportType()
		var f float64
		switch t.Kind() {
		case reflect.Float64:
			f = val.ToFloat()
		case reflect.Int64:
			f = float64(val.ToInteger())
		default:
			s := fmt.Sprintf("unexpected type for 'weight': got %s, expected Number", util.JsType(val))
			panic(n.m.vm.ToValue(s))
		}
		old := n.origNode.Weight
		n.origNode.Weight = f
		n.restore = append(n.restore, func() {
			n.origNode.Weight = old
		})
	case "dependsOn":
		if val.ExportType().Kind() != reflect.Slice {
			s := fmt.Sprintf("unexpected type for 'dependsOn': got %s, expected Array", util.JsType(val))
			panic(n.m.vm.ToValue(s))
		}
		arr := val.ToObject(n.m.vm)
		length := int(arr.Get("length").ToInteger())
		old := n.origNode.DependsOn
		for i := 0; i < length; i++ {
			item := arr.Get(strconv.Itoa(i))
			if attr, ok := item.Export().(string); !ok {
				s := fmt.Sprintf("unexpected type for 'dependsOn[%d]': got %s, expected Array", i, util.JsType(val))
				panic(n.m.vm.ToValue(s))
			} else {
				n.origNode.DependsOn = append(n.origNode.DependsOn, attr)
			}
		}
		n.restore = append(n.restore, func() {
			n.origNode.Attributes = old
		})
		return true
	case "children":
		if val.ExportType().Kind() != reflect.Slice {
			s := fmt.Sprintf("unexpected type for 'children': got %s, expected Array", util.JsType(val))
			panic(n.m.vm.ToValue(s))
		}
		arr := val.ToObject(n.m.vm)
		length := int(arr.Get("length").ToInteger())
		old := n.origNode.Children
		for i := 0; i < length; i++ {
			item := arr.Get(strconv.Itoa(i))
			n.origNode.Children = append(n.origNode.Children, convertToNode(item, n.m))
		}
		n.restore = append(n.restore, func() {
			n.origNode.Children = old
		})
		return true
	case "fake":
		f, ok := goja.AssertFunction(val)
		if !ok {
			s := fmt.Sprintf("unexpected type for 'fake': got %s, expected function", util.JsType(val))
			panic(n.m.vm.ToValue(s))
		}
		old := n.origNode.Fake
		n.origNode.Fake = func(r *generator.Request) (interface{}, error) {
			v, err := n.m.loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
				param := n.m.vm.ToValue(r)
				return f(goja.Undefined(), param)
			})
			if err != nil {
				return nil, err
			}
			return v.Export(), err
		}
		n.restore = append(n.restore, func() {
			n.origNode.Fake = old
		})
		return true
	}
	return false
}

func (n *Node) Delete(_ string) bool {
	return false
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
	return len(a.n.origNode.Attributes)
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
