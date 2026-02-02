package sortedmap

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// LinkedHashMap defines the iteration ordering by the order
// in which keys were inserted into the map
type LinkedHashMap[K comparable, V any] struct {
	KeyNormalizer func(K) K

	pairs map[interface{}]*pair[K, V]
	list  *list.List
}

type pair[K comparable, V any] struct {
	key     K
	value   V
	element *list.Element
}

func NewLinkedHashMap() *LinkedHashMap[string, interface{}] {
	return &LinkedHashMap[string, interface{}]{}
}

func (m *LinkedHashMap[K, V]) Set(key K, value V) {
	nk := key
	if m.KeyNormalizer != nil {
		nk = m.KeyNormalizer(key)
	}
	m.ensureInit()
	p, ok := m.pairs[nk]
	if !ok {
		p = &pair[K, V]{key: nk, value: value}
		m.pairs[nk] = p
		p.element = m.list.PushBack(p)
	} else {
		p.value = value
	}
}

func (m *LinkedHashMap[K, V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.pairs)
}

func (m *LinkedHashMap[K, V]) Get(key K) (V, bool) {
	if m.pairs != nil {
		nk := key
		if m.KeyNormalizer != nil {
			nk = m.KeyNormalizer(key)
		}
		p, ok := m.pairs[nk]
		if ok {
			return p.value, true
		}
	}
	return *new(V), false
}

func (m *LinkedHashMap[K, V]) Del(key K) {
	if m.pairs == nil {
		return
	}
	nk := key
	if m.KeyNormalizer != nil {
		nk = m.KeyNormalizer(key)
	}
	p, ok := m.pairs[nk]
	if !ok {
		return
	}
	delete(m.pairs, nk)
	m.list.Remove(p.element)
}

func (m *LinkedHashMap[K, V]) Lookup(key K) V {
	nk := key
	if m.KeyNormalizer != nil {
		nk = m.KeyNormalizer(key)
	}
	v, _ := m.Get(nk)
	return v
}

func (m *LinkedHashMap[K, V]) Iter() *Iterator[K, V] {
	if m == nil || m.list == nil {
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

func (m *LinkedHashMap[K, V]) ToMap() map[K]V {
	result := map[K]V{}
	for it := m.Iter(); it.Next(); {
		result[it.Key()] = it.Value()
	}
	return result
}

func (m *LinkedHashMap[K, V]) Resolve(token string) (interface{}, error) {
	for k, v := range m.pairs {
		if k == token {
			return v.value, nil
		}
	}
	return nil, fmt.Errorf("unable to resolve %v", token)
}

func (m *LinkedHashMap[K, V]) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); ok && delim != '{' {
		return fmt.Errorf("expected map, got %s", token)
	}
	for {
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}

		v := *new(V)
		err = dec.Decode(&v)
		if err != nil {
			return err
		}
		m.Set(token.(K), v)
	}
}

func (m *LinkedHashMap[K, V]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected map, got %v", value.Tag)
	}
	for i := 0; i < len(value.Content); i += 2 {
		var key any
		err := value.Content[i].Decode(&key)
		if err != nil {
			return err
		}
		v := *new(V)
		err = value.Content[i+1].Decode(&v)
		if err != nil {
			return err
		}
		m.Set(key.(K), v)
	}
	return nil
}
