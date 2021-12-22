package sortedmap

import "container/list"

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

func (m *LinkedHashMap) Get(key interface{}) (interface{}, bool) {
	if m.pairs != nil {
		if p, ok := m.pairs[key]; ok {
			return p.value, true
		}
	}
	return nil, false
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
