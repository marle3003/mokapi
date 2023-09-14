package sortedmap

import "container/list"

type Iterator[K comparable, V any] struct {
	next    *list.Element
	current *list.Element
}

func (i *Iterator[K, V]) Next() bool {
	if i.next == nil {
		return false
	}
	i.current = i.next
	i.next = i.next.Next()
	return true
}

func (i *Iterator[K, V]) Item() (K, V) {
	if i.current == nil {
		panic("current is nil")
	}
	p := i.current.Value.(*pair[K, V])
	return p.key, p.value
}

func (i *Iterator[K, V]) Key() K {
	k, _ := i.Item()
	return k
}

func (i *Iterator[K, V]) Value() V {
	_, v := i.Item()
	return v
}
