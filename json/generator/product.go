package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func ProductTree() *Tree {
	return &Tree{
		Name: "Product",
		nodes: []*Tree{
			ProductName(),
			ProductDescription(),
			ProductMaterial(),
			ProductCategory(),
		},
	}
}

func ProductName() *Tree {
	return &Tree{
		Name: "ProductName",
		Test: func(r *Request) bool {
			if !(r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == "") && !r.Schema.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return r.matchLast([]string{"product", "name"}, true) || last == "productname"
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
			if !(r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == "") && !r.Schema.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return r.matchLast([]string{"product", "description"}, true) || last == "productdescription"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductDescription(), nil
		},
	}
}

func ProductMaterial() *Tree {
	return &Tree{
		Name: "ProductMaterial",
		Test: func(r *Request) bool {
			if !(r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == "") && !r.Schema.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return r.matchLast([]string{"product", "material"}, true) || last == "material"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductMaterial(), nil
		},
	}
}

func ProductCategory() *Tree {
	return &Tree{
		Name: "ProductMaterial",
		Test: func(r *Request) bool {
			if !(r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == "") && !r.Schema.IsAny() {
				return false
			}
			last := strings.ToLower(r.LastName())
			return r.matchLast([]string{"product", "category"}, true) || last == "category"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.ProductCategory(), nil
		},
	}
}
