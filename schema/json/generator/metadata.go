package generator

func metadata() []*Node {
	return []*Node{
		{
			Name:       "tag",
			Attributes: []string{"tag", "tagName", "tagsName"},
			Fake:       fakeName,
		},
	}
}
