package metrics

import (
	"fmt"
	"hash/fnv"
	"strings"
)

type Metric interface {
	Info() *Info
	Collect(ch chan<- Metric)
}

type Label struct {
	Name  string
	Value string
}

type Info struct {
	Namespace string
	Name      string
	labels    []*Label
}

func (i *Info) String() string {
	if len(i.labels) == 0 {
		return i.FQName()
	}

	labels := make([]string, 0, len(i.labels))
	for _, l := range i.labels {
		labels = append(labels, fmt.Sprintf("%v=\"%v\"", l.Name, l.Value))
	}
	return fmt.Sprintf("%v{%v}", i.FQName(), strings.Join(labels, ","))
}

func (i *Info) getLabel(name string) (*Label, bool) {
	for _, l := range i.labels {
		if l.Name == name {
			return l, true
		}
	}
	return nil, false
}

func (i *Info) FQName() string {
	if len(i.Namespace) == 0 {
		return i.Name
	}
	return fmt.Sprintf("%v_%v", i.Namespace, i.Name)
}

func (i *Info) Match(query *Query) bool {
	if len(query.FQName) > 0 && query.FQName != i.FQName() {
		return false
	}
	if len(query.Namespace) > 0 && query.Namespace != i.Namespace {
		return false
	}
	if len(query.Name) > 0 && query.Name != i.Name {
		return false
	}
	for _, ql := range query.Labels {
		if l, ok := i.getLabel(ql.Name); !ok || l.Value != ql.Value {
			return false
		}
	}
	return true
}

type Options func(o *options)

type options struct {
	namespace  string
	name       string
	labels     []*Label
	labelNames []string
}

func WithName(name string) Options {
	return func(o *options) {
		o.name = name
	}
}

func WithLabelNames(names ...string) Options {
	return func(o *options) {
		o.labelNames = names
	}
}

func WithFQName(namespace, name string) Options {
	return func(o *options) {
		o.namespace = namespace
		o.name = name
	}
}

func WithLabels(labels ...*Label) Options {
	return func(o *options) {
		o.labels = append(o.labels, labels...)
	}
}

func hash(values []string) uint32 {
	s := strings.Join(values, "_")

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
