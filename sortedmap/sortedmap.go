package sortedmap

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"strings"
)

// LinkedHashMap defines the iteration ordering by the order
// in which keys were inserted into the map
type LinkedHashMap[K comparable, V any] struct {
	pairs map[interface{}]*pair[K, V]
	list  *list.List
}

type pair[K comparable, V any] struct {
	key   K
	value V
}

func NewLinkedHashMap() *LinkedHashMap[string, interface{}] {
	return &LinkedHashMap[string, interface{}]{}
}

//func FromMap(m interface{}) *LinkedHashMap[string, interface{}] {
//	sm := NewLinkedHashMap()
//	v := reflect.ValueOf(m)
//	for _, k := range v.MapKeys() {
//		sm.Set(k.Convert(reflect.), v.MapIndex(k).Interface())
//	}
//	return sm
//}

func (m *LinkedHashMap[K, V]) Set(key K, value V) {
	m.ensureInit()
	p, ok := m.pairs[key]
	if !ok {
		p = &pair[K, V]{key: key, value: value}
		m.pairs[key] = p
		m.list.PushBack(p)
	} else {
		p.value = value
	}
}

func (m *LinkedHashMap[K, V]) Len() int {
	return len(m.pairs)
}

func (m *LinkedHashMap[K, V]) Get(key K) V {
	if m.pairs != nil {
		p, ok := m.pairs[key]
		if ok {
			return p.value
		}
	}
	return *new(V)
}

func (m *LinkedHashMap[K, V]) Iter() *Iterator[K, V] {
	if m.list == nil {
		return &Iterator[K, V]{}
	}
	return &Iterator[K, V]{next: m.list.Front()}
}

func (m *LinkedHashMap[K, V]) Keys() []K {
	v := make([]K, 0, len(m.pairs))
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Key())
	}
	return v
}

func (m *LinkedHashMap[K, V]) Values() []V {
	v := make([]V, 0, len(m.pairs))
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Value())
	}
	return v
}

func (m *LinkedHashMap[K, V]) ensureInit() {
	if m.pairs == nil {
		m.pairs = make(map[interface{}]*pair[K, V])
		m.list = list.New()
	}
}

//func (m *LinkedHashMap[K, V]) Resolve(name K) (interface{}, error) {
//	if name == "*" {
//		return m.Values(), nil
//	}
//	v, ok := m.pairs[name]
//	if ok {
//		return v.value, nil
//	}
//	return nil, fmt.Errorf("undefined field %q", name)
//}

func (m *LinkedHashMap[K, V]) String() string {
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

func (m *LinkedHashMap[K, V]) Merge(o *LinkedHashMap[K, V]) {
	for it := o.Iter(); it.Next(); {
		m.Set(it.Key(), it.Value())
	}
}

func (m *LinkedHashMap[K, V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for it := m.Iter(); it.Next(); {
		if buf.Len() > 1 {
			buf.WriteString(",")
		}
		fmt.Fprintf(&buf, "\"%v\":", it.Key())
		value, err := json.Marshal(it.Value())
		if err != nil {
			return nil, err
		}
		buf.Write(value)
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}
