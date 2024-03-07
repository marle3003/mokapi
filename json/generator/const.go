package generator

func Const() *Tree {
	return &Tree{
		Name: "Example",
		compare: func(r *Request) bool {
			return r.Schema != nil && r.Schema.Const != nil
		},
		resolve: func(r *Request) (interface{}, error) {
			return r.Schema.Const, nil
		},
	}
}
