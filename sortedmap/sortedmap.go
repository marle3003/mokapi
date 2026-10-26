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

	values map[K]V
	keys   []K
	stale  int
}

type pair[K comparable, V any] struct {
	key     K
	value   V
	element *list.Element
}

func NewLinkedHashMap() *LinkedHashMap[string, any] {
	return &LinkedHashMap[string, any]{}
}

func (m *LinkedHashMap[K, V]) normalize(key K) K {
	if m.KeyNormalizer != nil {
		return m.KeyNormalizer(key)
	}
	return key
}

func (m *LinkedHashMap[K, V]) compact() {
	keys := m.keys[:0]
	for _, k := range m.keys {
		if _, ok := m.values[k]; ok {
			keys = append(keys, k)
		}
	}
	var zero K
	for i := len(keys); i < len(m.keys); i++ {
		m.keys[i] = zero
	}
	m.keys = keys
	m.stale = 0
}

func (m *LinkedHashMap[K, V]) Set(key K, value V) {
	nk := m.normalize(key)
	m.ensureInit()
	if _, ok := m.values[nk]; !ok {
		if m.stale > 0 {
			m.compact()
		}
		m.keys = append(m.keys, nk)
	}
	m.values[nk] = value
}

func (m *LinkedHashMap[K, V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

func (m *LinkedHashMap[K, V]) Get(key K) (V, bool) {
	if m != nil && m.values != nil {
		v, ok := m.values[m.normalize(key)]
		return v, ok
	}
	return *new(V), false
}

func (m *LinkedHashMap[K, V]) Del(key K) {
	if m.values == nil {
		return
	}
	nk := m.normalize(key)
	if _, ok := m.values[nk]; !ok {
		return
	}
	delete(m.values, nk)
	m.stale++
	if m.stale > 32 && m.stale > len(m.keys)/2 {
		m.compact()
	}
}

func (m *LinkedHashMap[K, V]) Lookup(key K) V {
	v, _ := m.Get(key)
	return v
}

func (m *LinkedHashMap[K, V]) Iter() *Iterator[K, V] {
	if m == nil {
		return &Iterator[K, V]{}
	}
	return &Iterator[K, V]{m: m, i: -1}
}

func (m *LinkedHashMap[K, V]) Keys() []K {
	v := make([]K, 0, len(m.keys))
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Key())
	}
	return v
}

func (m *LinkedHashMap[K, V]) Values() []V {
	v := make([]V, 0, m.Len())
	for it := m.Iter(); it.Next(); {
		v = append(v, it.Value())
	}
	return v
}

func (m *LinkedHashMap[K, V]) ensureInit() {
	if m.values == nil {
		m.values = make(map[K]V)
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
		_, _ = fmt.Fprintf(&buf, "\"%v\":", it.Key())
		value, err := json.Marshal(it.Value())
		if err != nil {
			return nil, err
		}
		buf.Write(value)
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}

func (m *LinkedHashMap[K, V]) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}

	for it := m.Iter(); it.Next(); {
		keyNode := &yaml.Node{}
		if err := keyNode.Encode(it.Key()); err != nil {
			return nil, err
		}

		valNode := &yaml.Node{}
		if err := valNode.Encode(it.Value()); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valNode)
	}

	return node, nil
}

// ToMap returns a copy of the entries as a plain map.
func (m *LinkedHashMap[K, V]) ToMap() map[K]V {
	result := make(map[K]V, m.Len())
	for it := m.Iter(); it.Next(); {
		result[it.Key()] = it.Value()
	}
	return result
}

func (m *LinkedHashMap[K, V]) Map() map[K]V {
	if m.values == nil {
		return map[K]V{}
	}
	return m.values
}

func (m *LinkedHashMap[K, V]) Resolve(token string) (interface{}, error) {
	for k, v := range m.values {
		if any(k) == any(token) {
			return v, nil
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
