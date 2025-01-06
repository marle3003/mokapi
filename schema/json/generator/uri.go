package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"strings"
)

func Uri() *Tree {
	return &Tree{
		Name: "URIs",
		Nodes: []*Tree{
			UriList(),
			SingleUri(),
		},
	}
}

func SingleUri() *Tree {
	return &Tree{
		Name: "URI",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				name := strings.ToLower(p.Name)
				if name == "uri" || name == "url" {
					return p.Schema.IsAnyString() || p.Schema.IsAny()
				}
				if strings.HasSuffix(name, "uri") || strings.HasSuffix(name, "url") {
					return p.Schema.IsAnyString()
				}

				return false
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.URL(), nil
		},
	}
}

func UriList() *Tree {
	return &Tree{
		Name: "URI-List",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				name := strings.ToLower(p.Name)
				if name == "uris" || name == "urls" {
					return p.Schema.IsArray() || p.Schema.IsAny()
				}
				if strings.HasSuffix(name, "uris") || strings.HasSuffix(name, "urls") {
					return p.Schema.IsArray()
				}

				return false
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			last := r.Last()
			s := last.Schema
			if s.IsAny() {
				s = &schema.Ref{Value: &schema.Schema{Type: []string{"array"}}}
			}

			next := r.With()
			next.Path = Path{&PathElement{
				Name:   strings.TrimSuffix(last.Name, "s"),
				Schema: s,
			}}
			return r.g.tree.Resolve(next)
		},
	}
}
