package generator

func metadata() []*Node {
	return []*Node{
		{
			Name: "tag",
			Fake: fakeName,
		},
		{
			Name: "tags",
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeName,
				},
			},
		},
	}
}
