package generator

func newKeyNode() *Node {
	return &Node{Name: "key", Fake: fakeKey}
}

func fakeKey(r *Request) (interface{}, error) {
	s := r.Schema
	if s.IsString() {
		if s.Pattern != "" {
			return fakePattern(r)
		}
		return fakeId(r)
	}

	return nil, NotSupported
}
