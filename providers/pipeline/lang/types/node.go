package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/encoding"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
	"strings"
)

type Node struct {
	attributes *Expando
	children   *Array
	name       string
	content    string
}

func NewNode(name string) *Node {
	return &Node{name: name, attributes: NewExpando(), children: NewArray()}
}

func ParseNode(node *encoding.XmlNode) *Node {
	n := NewNode(node.Name)
	n.content = node.Content
	for k, v := range node.Attributes {
		n.attributes.value[k] = NewString(v)
	}

	for _, c := range node.Children {
		n.children.Add(ParseNode(c))
	}
	return n
}

func (n *Node) Index(index int) (Object, error) {
	return n.children.Index(index)
}

func (n *Node) String() string {
	sb := strings.Builder{}
	hasName := len(n.name) > 0
	if hasName {
		sb.WriteString(fmt.Sprintf("<%v", n.name))
	}
	for k, v := range n.attributes.value {
		sb.WriteString(fmt.Sprintf(" %v=\"%v\"", k, v))
	}
	if len(n.children.value) == 0 && len(n.content) == 0 && hasName {
		sb.WriteString(" />")
	} else {
		if hasName {
			sb.WriteString(">")
		}

		for _, c := range n.children.value {
			sb.WriteString(c.String())
		}

		sb.WriteString(n.content)

		if hasName {
			sb.WriteString(fmt.Sprintf("</%v>", n.name))
		}
	}

	return sb.String()
}

func (n *Node) Add(obj Object) {
	if node, ok := obj.(*Node); ok {
		n.children.Add(node)
	}
}

func (n *Node) Set(o Object) error {
	if v, isArray := o.(*Array); isArray {
		n.children = v
		return nil
	} else {
		return errors.Errorf("type '%v' can not be set to node", o.GetType())
	}
}

func (n *Node) Children() *Array {
	a := NewArray()
	for _, c := range n.children.value {
		a.Add(c)
	}
	return a
}

func (n *Node) Find(match Predicate) (Object, error) {
	for _, item := range n.children.value {
		if matches, err := match(item); err == nil {
			if matches {
				return item, nil
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (n *Node) FindAll(match Predicate) (*Array, error) {
	result := NewArray()
	for _, c := range n.children.value {
		if matches, err := match(c); err == nil {
			if matches {
				result.Add(c)
			}
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (n *Node) Elem() interface{} {
	xml := &encoding.XmlNode{
		Name:    n.name,
		Content: n.content,
	}
	if attr, ok := n.attributes.Elem().(map[string]string); ok {
		for k, v := range attr {
			xml.Attributes[k] = v
		}
	}
	if children, ok := n.children.Elem().([]interface{}); ok {
		for _, i := range children {
			if c, ok := i.(*Node); ok {
				xml.Children = append(xml.Children, c.Elem().(*encoding.XmlNode))
			}
		}
	}

	return xml
}

func (n *Node) GetType() reflect.Type {
	return reflect.TypeOf(n)
}

func (n *Node) InvokeOp(op token.Token, o Object) (Object, error) {
	return NewString(n.content).InvokeOp(op, o)
}

func (n *Node) Name() *String {
	return NewString(n.name)
}

func (n *Node) Text() *String {
	var sb strings.Builder
	for _, i := range n.children.value {
		if c, ok := i.(*Node); ok {
			sb.WriteString(c.Text().String())
		}
	}
	sb.WriteString(n.content)
	return NewString(sb.String())
}

func (n *Node) GetField(name string) (Object, error) {
	if strings.HasPrefix(name, "@") {
		name = name[1:]
		return n.attributes.GetField(name)
	} else {
		for _, i := range n.children.value {
			if c, ok := i.(*Node); ok && c.name == name {
				return c, nil
			} else if !ok {
				return nil, errors.Errorf("unexpected type of child node: %v", reflect.TypeOf(i))
			}
		}
	}

	switch strings.ToLower(name) {
	case "name":
		return n.Name(), nil
	case "text":
		return n.Text(), nil
	case "children":
		return n.children, nil
	case "attributes":
		return n.children, nil
	}

	return getField(n, name)
}

func (n *Node) HasField(name string) bool {
	switch strings.ToLower(name) {
	case "name", "text", "children":
		return true
	}
	return hasField(n, name)
}

func (n *Node) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(n, name, args)
}

func (n *Node) SetField(name string, value Object) error {
	if strings.HasPrefix(name, "@") {
		name = name[1:]
		n.attributes.SetField(name, value)
	}

	switch strings.ToLower(name) {
	case "name":
		n.name = value.String()
	case "text":
		n.content = value.String()
	case "children":
		if array, ok := value.(*Array); ok {
			n.children = array
		} else {
			return errors.Errorf("unexpected type of value: %v", reflect.TypeOf(value))
		}
	}

	return setField(n, name, value)
}
