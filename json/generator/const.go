package generator

func Const() *Tree {
	return &Tree{
		Name: "Example",
		Test: func(r *Request) bool {
			return r.Schema != nil && r.Schema.Const != nil
		},
		Fake: func(r *Request) (interface{}, error) {
			return r.Schema.Const, nil
		},
	}
}
