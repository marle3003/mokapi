package openapi

// cloneMap copies a map and its values. A nil map stays nil.
func cloneMap[K comparable, V any](m map[K]V, clone func(V) V) map[K]V {
	if m == nil {
		return nil
	}
	c := make(map[K]V, len(m))
	for k, v := range m {
		c[k] = clone(v)
	}
	return c
}

// cloneSlice copies a slice and its elements. A nil slice stays nil.
func cloneSlice[V any](s []V, clone func(V) V) []V {
	if s == nil {
		return nil
	}
	c := make([]V, len(s))
	for i, v := range s {
		c[i] = clone(v)
	}
	return c
}

// clonePointer copies the value p points to. A nil pointer stays nil.
func clonePointer[T any](p *T) *T {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}
