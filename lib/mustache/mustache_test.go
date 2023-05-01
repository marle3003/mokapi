package mustache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRender(t *testing.T) {
	testcases := []struct {
		name     string
		template string
		scope    interface{}
		expected string
	}{
		{
			name:     "raw text",
			template: "<h1>Foo</h1>",
			scope:    nil,
			expected: "<h1>Foo</h1>",
		},
		{
			name:     "with a struct",
			template: "<h1>{{ foo }}</h1>",
			scope:    struct{ Foo string }{Foo: "Bar"},
			expected: "<h1>Bar</h1>",
		},
		{
			name:     "with a nested struct",
			template: "<h1>{{ foo.name }}</h1>",
			scope:    struct{ Foo interface{} }{Foo: struct{ Name string }{Name: "Bob"}},
			expected: "<h1>Bob</h1>",
		},
		{
			name:     "with a map",
			template: "<h1>{{ foo }}</h1>",
			scope:    map[string]interface{}{"foo": "Bar"},
			expected: "<h1>Bar</h1>",
		},
		{
			name:     "with a map without whitespace",
			template: "<h1>{{foo}}</h1>",
			scope:    map[string]interface{}{"foo": "Bar"},
			expected: "<h1>Bar</h1>",
		},
		{
			name:     "with a map with whitespace",
			template: "<h1>{{        foo              }}</h1>",
			scope:    map[string]interface{}{"foo": "Bar"},
			expected: "<h1>Bar</h1>",
		},
		{
			name:     "with a nested map",
			template: "<h1>{{ foo.name }}</h1>",
			scope: map[string]interface{}{"foo": map[string]interface{}{
				"name": "Bob",
				"mail": "bob@mokapi.io",
			}},
			expected: "<h1>Bob</h1>",
		},
		{
			name:     "root",
			template: "<h1>{{ . }}</h1>",
			scope:    "Bar",
			expected: "<h1>Bar</h1>",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, err := Render(tc.template, tc.scope)
			require.NoError(t, err)
			require.Equal(t, tc.expected, s)
		})
	}
}

func TestRender_Error(t *testing.T) {
	_, err := Render("{{ foo }}", map[interface{}]interface{}{12: "foo"})
	require.Error(t, err)
}
