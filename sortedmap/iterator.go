package sortedmap

type Iterator[K comparable, V any] struct {
	m   *LinkedHashMap[K, V]
	i   int
	key K
	val V
}

func (i *Iterator[K, V]) Next() bool {
	if i.m == nil {
		return false
	}
	for i.i+1 < len(i.m.keys) {
		i.i++
		k := i.m.keys[i.i]
		if v, ok := i.m.values[k]; ok {
			i.key, i.val = k, v
			return true
		}
	}
	return false
}

func (i *Iterator[K, V]) Item() (K, V) {
	if i.m == nil || i.i < 0 {
		panic("current is nil")
	}
	return i.key, i.val
}

func (i *Iterator[K, V]) Key() K {
	return i.key
}

func (i *Iterator[K, V]) Value() V {
	return i.val
}
