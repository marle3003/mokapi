package generator

import (
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShallowestLevels(t *testing.T) {
	str := func() *schema.Schema { return schematest.New("string") }
	obj := func(opts ...schematest.SchemaOptions) *schema.Schema { return schematest.New("object", opts...) }
	withId := func() *schema.Schema { return obj(schematest.WithProperty("id", str())) }
	idOnly := []string{"id"}

	testcases := []struct {
		name   string
		schema *schema.Schema
		names  []string
		expect map[string]int
	}{
		{
			name:   "root object",
			schema: obj(schematest.WithProperty("id", str())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "wrapper",
			schema: obj(schematest.WithProperty("data", withId())),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name:   "root array",
			schema: schematest.New("array", schematest.WithItems("object", schematest.WithProperty("id", str()))),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name: "items of a wrapper",
			schema: obj(schematest.WithProperty("items",
				schematest.New("array", schematest.WithItems("object", schematest.WithProperty("id", str()))))),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name:   "root property wins over a nested one",
			schema: obj(schematest.WithProperty("a", withId()), schematest.WithProperty("id", str())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name: "siblings",
			schema: obj(
				schematest.WithProperty("billing", withId()),
				schematest.WithProperty("shipping", withId())),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name: "uneven depth",
			schema: obj(
				schematest.WithProperty("a", withId()),
				schematest.WithProperty("b", obj(schematest.WithProperty("c", withId())))),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name:   "required name without a property",
			schema: obj(schematest.WithRequired("id")),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "oneOf variant belongs to the level of the object",
			schema: obj(schematest.WithOneOf(withId(), obj(schematest.WithProperty("code", str())))),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "anyOf variant belongs to the level of the object",
			schema: obj(schematest.Any(withId())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "allOf variant belongs to the level of the object",
			schema: obj(schematest.WithAllOf(withId())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name: "variant of a nested object",
			schema: obj(schematest.WithProperty("data",
				obj(schematest.WithOneOf(withId())))),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name: "variant property is shallower than a nested property",
			schema: obj(
				schematest.WithProperty("a", withId()),
				schematest.WithAllOf(withId())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "conditional schema",
			schema: obj(schematest.WithIf(obj()), schematest.WithThen(withId())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "dependent schema",
			schema: obj(schematest.WithDependentSchemas("a", withId())),
			names:  idOnly,
			expect: map[string]int{"id": 1},
		},
		{
			name:   "prefix items and contains",
			schema: schematest.New("array", schematest.WithPrefixItems(withId()), schematest.WithContains(obj(schematest.WithProperty("code", str())))),
			names:  []string{"id", "code"},
			expect: map[string]int{"id": 1, "code": 1},
		},
		{
			name:   "pattern and additional properties are ignored",
			schema: obj(schematest.WithPatternProperty("^id$", withId()), schematest.WithAdditionalProperties(withId())),
			names:  idOnly,
			expect: map[string]int{"id": 2},
		},
		{
			name:   "name not found",
			schema: obj(schematest.WithProperty("code", str())),
			names:  idOnly,
			expect: map[string]int{},
		},
		{
			name:   "primitive root",
			schema: str(),
			names:  idOnly,
			expect: map[string]int{},
		},
		{
			name:   "several names",
			schema: obj(schematest.WithProperty("id", str()), schematest.WithProperty("data", obj(schematest.WithProperty("status", str())))),
			names:  []string{"id", "status"},
			expect: map[string]int{"id": 1, "status": 2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, shallowestLevels(tc.schema, tc.names))
		})
	}

	t.Run("recursive schema terminates", func(t *testing.T) {
		node := obj(schematest.WithProperty("id", str()))
		node.Properties.Set("child", node)

		require.Equal(t, map[string]int{"id": 1}, shallowestLevels(node, idOnly))
		require.Equal(t, map[string]int{}, shallowestLevels(node, []string{"missing"}))
	})

	t.Run("recursive reference terminates", func(t *testing.T) {
		const ref = "#/components/schemas/Node"
		// a resolved reference may be a copy of the schema it points to
		node := obj(schematest.WithProperty("id", str()))
		node.Ref = ref
		copyOfNode := obj(schematest.WithProperty("id", str()))
		copyOfNode.Ref = ref
		node.Properties.Set("child", copyOfNode)
		copyOfNode.Properties.Set("child", node)

		require.Equal(t, map[string]int{"id": 1}, shallowestLevels(node, idOnly))
		require.Equal(t, map[string]int{}, shallowestLevels(node, []string{"missing"}))
	})
}
