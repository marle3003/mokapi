package sortedmap

import (
	"container/list"
	"fmt"
	"strings"
)

// LinkedHashMap defines the iteration ordering by the order
// in which keys were inserted into the map
type LinkedHashMap struct {
	pairs map[interface{}]*pair
	list  *list.List
}

type pair struct {
	key   interface{}
	value interface{}
}

func NewLinkedHashMap() *LinkedHashMap {
	return &LinkedHashMap{}
}

func (m *LinkedHashMap) Set(key, value interface{}) {
	m.ensureInit()
	p, ok := m.pairs[key]
	if !ok {
		p = &pair{key: key, value: value}
		m.pairs[key] = p
		m.list.PushBack(p)
	} else {
		p.value = value
	}
}

func (m *LinkedHashMap) Len() int {
	return len(m.pairs)
}

func (m *LinkedHashMap) Get(key interface{}) interface{} {
	if m.pairs != nil {
		p, ok := m.pairs[key]
		if ok {
			return p.value
		}
	}
	return nil
}

func (m *LinkedHashMap) Iter() *Iterator {
	if m.list == nil {
		return &Iterator{}
	}
	return &Iterator{next: m.list.Front()}
}

func (m *LinkedHashMap) Keys() []interface{} {
	v := make([]interface{}, 0, len(m.pairs))
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Key())
	}
	return v
}

func (m *LinkedHashMap) Values() []interface{} {
	v := make([]interface{}, 0, len(m.pairs))
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Value())
	}
	return v
}

func (m *LinkedHashMap) ensureInit() {
	if m.pairs == nil {
		m.pairs = make(map[interface{}]*pair)
		m.list = list.New()
	}
}

func (m *LinkedHashMap) Resolve(name interface{}) (interface{}, error) {
	if name == "*" {
		return m.Values(), nil
	}
	v, ok := m.pairs[name]
	if ok {
		return v.value, nil
	}
	return nil, fmt.Errorf("undefined field %q", name)
}

func (m *LinkedHashMap) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	for it := m.Iter(); it.Next(); {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v: %v", it.Key(), it.Value()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *LinkedHashMap) Merge(o *LinkedHashMap) {
	for it := o.Iter(); it.Next(); {
		m.Set(it.Key(), it.Value())
	}
}
