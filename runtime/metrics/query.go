package metrics

type Query struct {
	FQName    string
	Namespace string
	Name      string
	Labels    []*Label
}

type QueryOptions func(q *Query)

func NewQuery(options ...QueryOptions) *Query {
	q := &Query{}
	for _, o := range options {
		o(q)
	}
	return q
}

func ByFQName(s string) QueryOptions {
	return func(q *Query) {
		q.FQName = s
	}
}

func ByNamespace(s string) QueryOptions {
	return func(q *Query) {
		q.Namespace = s
	}
}

func ByName(s string) QueryOptions {
	return func(q *Query) {
		q.Name = s
	}
}

func ByLabel(name, value string) QueryOptions {
	return func(q *Query) {
		q.Labels = append(q.Labels, &Label{Name: name, Value: value})
	}
}
