package metrics

type Query struct {
	FQName    string
	Namespace string
	Name      string
	Labels    []*Label
}

type QueryFunc func(q *Query)

func ByFQName(s string) QueryFunc {
	return func(q *Query) {
		q.FQName = s
	}
}

func ByNamespace(s string) QueryFunc {
	return func(q *Query) {
		q.Namespace = s
	}
}

func ByName(s string) QueryFunc {
	return func(q *Query) {
		q.Name = s
	}
}

func ByLabel(name, value string) QueryFunc {
	return func(q *Query) {
		q.Labels = append(q.Labels, &Label{Name: name, Value: value})
	}
}
