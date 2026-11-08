package asyncapi3

import "mokapi/sortedmap"

// A YAML or JSON null entry (e.g. "channels: {orders: }") decodes to a nil
// pointer in maps and slices. Parse removes such entries, so that all later
// readers see them as missing instead of dereferencing nil.

func removeNilEntries[K comparable, V any](m map[K]*V) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
}

func removeNilElements[V any](s []*V) []*V {
	n := 0
	for _, v := range s {
		if v != nil {
			s[n] = v
			n++
		}
	}
	clear(s[n:])
	return s[:n]
}

func removeNilServers(m *sortedmap.LinkedHashMap[string, *ServerRef]) {
	if m == nil {
		return
	}
	for _, k := range m.Keys() {
		if m.Lookup(k) == nil {
			m.Del(k)
		}
	}
}
