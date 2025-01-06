package generator

import "strings"

type ComparerFunc func(p *PathElement) bool

func (f ComparerFunc) Compare(p *PathElement) bool {
	return f(p)
}

func Any() Comparer {
	return ComparerFunc(func(p *PathElement) bool {
		return true
	})
}

func NameIgnoreCase(args ...string) Comparer {
	return ComparerFunc(func(p *PathElement) bool {
		for _, s := range args {
			if strings.ToLower(s) == strings.ToLower(p.Name) {
				return true
			}
		}
		return false
	})
}

func IsSchemaAny() Comparer {
	return ComparerFunc(func(p *PathElement) bool {
		return p.Schema.IsAny()
	})
}
