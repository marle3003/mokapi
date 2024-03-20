package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Products() *Tree {
	return &Tree{
		Name: "Product",
		Test: func(r *Request) bool {
			last := r.LastName()
			return r.Path.MatchLast(NameIgnoreCase("product"), Any()) ||
				strings.HasPrefix(strings.ToLower(last), "product")
		},
		Nodes: []*Tree{
			ProductName(),
			ProductDescription(),
			Material(),
			ProductCategory(),
		},
	}
}

func ProductName() *Tree {
	return &Tree{
		Name: "ProductName",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if !(s.IsString() && s.Pattern == "" && s.Format == "") && !s.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return last == "name" || last == "productname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductName(), nil
		},
	}
}

func ProductDescription() *Tree {
	return &Tree{
		Name: "ProductDescription",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if !(s.IsString() && s.Pattern == "" && s.Format == "") && !s.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return last == "description" || last == "productdescription"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductDescription(), nil
		},
	}
}

func Material() *Tree {
	return &Tree{
		Name: "ProductMaterial",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if !(s.IsString() && s.Pattern == "" && s.Format == "") && !s.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return last == "material"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductMaterial(), nil
		},
	}
}

func ProductCategory() *Tree {
	return &Tree{
		Name: "ProductCategory",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			if !(s.IsString() && s.Pattern == "" && s.Format == "") && !s.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return last == "category"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductCategory(), nil
		},
	}
}
