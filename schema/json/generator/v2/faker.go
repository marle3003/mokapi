package v2

var (
	nullFaker = &faker{fake: func() (interface{}, error) {
		return nil, nil
	}}
)

type faker struct {
	fake fakeFunc
	node *Node
}

func newFakerWithFallback(n *Node, r *Request) *faker {
	if n == nil || n.Fake == nil {
		return &faker{fake: func() (interface{}, error) {
			return fakeBySchema(r)
		}}
	}
	return &faker{
		fake: func() (interface{}, error) {
			v, err := n.Fake(r)
			if err != nil {
				return nil, err
			}
			if v, err = validate(v, r); err != nil {
				return fakeBySchema(r)
			}
			return v, nil
		},
		node: n,
	}
}

func newFaker(f fakeFunc) *faker {
	return &faker{fake: f}
}
