package sortedmap

import "container/list"

type Iterator struct {
	next    *list.Element
	current *list.Element
}

func (i *Iterator) Next() bool {
	if i.next == nil {
		return false
	}
	i.current = i.next
	i.next = i.next.Next()
	return true
}

func (i *Iterator) Item() (interface{}, interface{}) {
	if i.current == nil {
		panic("current is nil")
	}
	p := i.current.Value.(*pair)
	return p.key, p.value
}

func (i *Iterator) Key() interface{} {
	k, _ := i.Item()
	return k
}

func (i *Iterator) Value() interface{} {
	_, v := i.Item()
	return v
}
