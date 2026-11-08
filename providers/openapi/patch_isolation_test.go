package openapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/providers/openapi"
	"mokapi/sortedmap"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// patchAreas split the maximal config by the parts of Config.Patch that adopt values
// from the patch config.
var patchAreas = []struct {
	name string
	only func(c *openapi.Config) *openapi.Config
}{
	{"info", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Info: c.Info} }},
	{"servers", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Servers: c.Servers} }},
	{"security", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Security: c.Security} }},
	{"externalDocs", func(c *openapi.Config) *openapi.Config { return &openapi.Config{ExternalDocs: c.ExternalDocs} }},
	{"tags", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Tags: c.Tags} }},
	{"components", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Components: c.Components} }},
	{"paths", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Paths: c.Paths} }},
	{"webhooks", func(c *openapi.Config) *openapi.Config { return &openapi.Config{Webhooks: c.Webhooks} }},
}

// TestConfig_Patch_DoesNotChangeInputs patches configs, changes the result and
// requires that the patch configs are the same as freshly built ones.
func TestConfig_Patch_DoesNotChangeInputs(t *testing.T) {
	for _, area := range patchAreas {
		t.Run(area.name, func(t *testing.T) {
			t.Run("patch into empty", func(t *testing.T) {
				a := area.only(maximalConfig(false))
				result := &openapi.Config{}
				result.Patch(a)

				requireNothingShared(t, result, a)
				changeEverything(result)

				require.Equal(t, area.only(maximalConfig(false)), a)
			})

			t.Run("patch into merged config", func(t *testing.T) {
				a := area.only(maximalConfig(true))
				b := area.only(maximalConfig(false))
				result := &openapi.Config{}
				result.Patch(a)
				result.Patch(b)

				requireNothingShared(t, result, a, b)
				changeEverything(result)

				require.Equal(t, area.only(maximalConfig(true)), a)
				require.Equal(t, area.only(maximalConfig(false)), b)
			})
		})
	}
}

// requireNothingShared fails if a pointer, map or slice of result is also reachable from one of the inputs.
// Values stored in interfaces (examples, enum values) are taken over by Patch and not compared;
// the config a reference was resolved from (dynamic.Config) is shared on purpose.
func requireNothingShared(t *testing.T, result any, inputs ...any) {
	t.Helper()
	input := map[memory]bool{}
	for _, in := range inputs {
		walkMemory(reflect.ValueOf(in), map[memory]bool{}, func(m memory) { input[m] = true })
	}
	walkMemory(reflect.ValueOf(result), map[memory]bool{}, func(m memory) {
		require.False(t, input[m], "result shares %v at %x with a patch config", m.typ, m.addr)
	})
}

type memory struct {
	addr uintptr
	typ  reflect.Type
}

func walkMemory(v reflect.Value, seen map[memory]bool, visit func(memory)) {
	var m memory
	switch v.Kind() {
	case reflect.Ptr, reflect.Map:
		if v.IsNil() || v.Type() == reflect.TypeOf(&dynamic.Config{}) {
			return
		}
		m = memory{v.Pointer(), v.Type()}
	case reflect.Slice:
		if v.Cap() == 0 {
			return
		}
		m = memory{v.Pointer(), v.Type()}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			walkMemory(v.Field(i), seen, visit)
		}
		return
	default:
		return
	}
	if seen[m] {
		return
	}
	seen[m] = true
	visit(m)

	switch v.Kind() {
	case reflect.Ptr:
		walkMemory(v.Elem(), seen, visit)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			walkMemory(v.Index(i), seen, visit)
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			walkMemory(it.Value(), seen, visit)
		}
	}
}

// changeEverything writes to every field, slice element and map entry reachable
// from v, including unexported ones, except values stored in interfaces.
func changeEverything(v any) {
	change(reflect.ValueOf(v), map[uintptr]bool{})
}

func change(v reflect.Value, seen map[uintptr]bool) {
	if v.Kind() != reflect.Ptr && !v.CanSet() && v.CanAddr() {
		v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] || v.Type() == reflect.TypeOf(&dynamic.Config{}) {
			return
		}
		seen[v.Pointer()] = true
		change(v.Elem(), seen)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			change(v.Field(i), seen)
		}
	case reflect.String:
		v.SetString(v.String() + "-changed")
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(v.Int() + 1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(v.Uint() + 1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(v.Float() + 1)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			change(v.Index(i), seen)
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}
		for _, key := range v.MapKeys() {
			entry := reflect.New(v.Type().Elem()).Elem()
			entry.Set(v.MapIndex(key))
			change(entry, seen)
			v.SetMapIndex(key, entry)
		}
		if v.Type().Key().Kind() == reflect.String {
			v.SetMapIndex(reflect.ValueOf("added").Convert(v.Type().Key()), reflect.Zero(v.Type().Elem()))
		}
	}
}

func TestConfig_Patch_ReferencedComponentIsNotChanged(t *testing.T) {
	// The response of the path item is the one of the component, as Parse resolves a $ref.
	build := func() (*openapi.Config, *openapi.Response) {
		shared := &openapi.Response{Description: "from file a"}
		responses := &openapi.Responses{}
		responses.Set("200", &openapi.ResponseRef{
			Reference: dynamic.Reference[*openapi.ResponseRef]{Ref: "#/components/responses/ok"},
			Value:     shared,
		})
		paths := &openapi.PathItems{}
		paths.Set("/foo", &openapi.PathRef{Value: &openapi.Path{Get: &openapi.Operation{Responses: responses}}})
		return &openapi.Config{
			Paths:      paths,
			Components: openapi.Components{Responses: openapi.ResponseBodies{"ok": {Value: shared}}},
		}, shared
	}
	a, shared := build()

	b := &openapi.Config{}
	responses := &openapi.Responses{}
	responses.Set("200", &openapi.ResponseRef{Value: &openapi.Response{Description: "from file b"}})
	paths := &openapi.PathItems{}
	paths.Set("/foo", &openapi.PathRef{Value: &openapi.Path{Get: &openapi.Operation{Responses: responses}}})
	b.Paths = paths

	result := &openapi.Config{}
	result.Patch(a)
	result.Patch(b)

	merged, _ := result.Paths.Get("/foo")
	res, _ := merged.Value.Get.Responses.Get("200")
	require.Equal(t, "from file b", res.Value.Description)
	require.Equal(t, "from file a", shared.Description)
	expected, _ := build()
	require.Equal(t, expected, a)
}

func TestConfig_Patch_IncompleteTarget(t *testing.T) {
	t.Run("oauth2 scheme without flows and flow without scopes", func(t *testing.T) {
		flow := func() map[string]*openapi.OAuth2Flow {
			return map[string]*openapi.OAuth2Flow{"implicit": {Scopes: map[string]string{"read": "read access"}}}
		}
		result := &openapi.Config{Components: openapi.Components{SecuritySchemes: openapi.SecuritySchemes{
			"oauth": &openapi.OAuth2SecurityScheme{Type: "oauth2"},
		}}}
		result.Patch(&openapi.Config{Components: openapi.Components{SecuritySchemes: openapi.SecuritySchemes{
			"oauth": &openapi.OAuth2SecurityScheme{Type: "oauth2", Flows: flow()},
		}}})
		require.Equal(t, flow(), result.Components.SecuritySchemes["oauth"].(*openapi.OAuth2SecurityScheme).Flows)

		result = &openapi.Config{Components: openapi.Components{SecuritySchemes: openapi.SecuritySchemes{
			"oauth": &openapi.OAuth2SecurityScheme{Type: "oauth2", Flows: map[string]*openapi.OAuth2Flow{"implicit": {}}},
		}}}
		result.Patch(&openapi.Config{Components: openapi.Components{SecuritySchemes: openapi.SecuritySchemes{
			"oauth": &openapi.OAuth2SecurityScheme{Type: "oauth2", Flows: flow()},
		}}})
		require.Equal(t, flow(), result.Components.SecuritySchemes["oauth"].(*openapi.OAuth2SecurityScheme).Flows)
	})

	t.Run("example without value", func(t *testing.T) {
		result := &openapi.Config{Components: openapi.Components{Examples: openapi.Examples{
			"e": {Value: &openapi.Example{Summary: "summary"}},
		}}}
		result.Patch(&openapi.Config{Components: openapi.Components{Examples: openapi.Examples{
			"e": {},
		}}})
		require.Equal(t, "summary", result.Components.Examples["e"].Value.Summary)
	})

	t.Run("parameter without value", func(t *testing.T) {
		p := &openapi.ParameterRef{Value: &openapi.Parameter{Name: "id"}}
		p.Patch(&openapi.ParameterRef{})
		require.Equal(t, "id", p.Value.Name)

		empty := &openapi.ParameterRef{}
		empty.Patch(&openapi.ParameterRef{})
		require.Nil(t, empty.Value)
	})

	t.Run("nil entries in the patch config", func(t *testing.T) {
		result := &openapi.Config{}
		require.NotPanics(t, func() {
			result.Patch(&openapi.Config{
				Components: openapi.Components{
					Parameters:      openapi.ComponentParameters{"p": nil},
					Headers:         openapi.Headers{"h": nil},
					Examples:        openapi.Examples{"e": nil},
					RequestBodies:   openapi.RequestBodies{"r": nil},
					SecuritySchemes: openapi.SecuritySchemes{"s": nil},
				},
				Paths: &openapi.PathItems{LinkedHashMap: sortedmap.LinkedHashMap[string, *openapi.PathRef]{}},
			})
		})
		require.Nil(t, result.Components.Parameters["p"])
		require.Nil(t, result.Components.Headers["h"])
		require.Nil(t, result.Components.Examples["e"])
		require.Nil(t, result.Components.RequestBodies["r"])
		require.Nil(t, result.Components.SecuritySchemes["s"])
	})
}
