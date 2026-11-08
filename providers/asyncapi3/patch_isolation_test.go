package asyncapi3

import (
	"mokapi/config/dynamic"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestMaximalConfig_SetsAllFields(t *testing.T) {
	requireAllFieldsSet(t, maximalConfig("a"))
}

func TestConfig_Patch_IntoEmptyConfigKeepsEverything(t *testing.T) {
	a := maximalConfig("a")
	result := &Config{}

	result.Patch(a)

	for name, ch := range result.Channels {
		require.Same(t, result, ch.Value.Config, "channel %s must belong to the merged config", name)
		ch.Value.Config = nil
	}
	require.Equal(t, maximalConfig("a"), result)
}

// TestConfig_Patch_DoesNotChangeInputs patches configs, changes the result and
// requires that the patch configs are the same as freshly built ones.
func TestConfig_Patch_DoesNotChangeInputs(t *testing.T) {
	testcases := []struct {
		name string
		keys []string
	}{
		{name: "patch into empty", keys: []string{"a"}},
		{name: "patch merges entries with the same key", keys: []string{"a", "a"}},
		{name: "patch adds entries with other keys", keys: []string{"a", "b"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var inputs []*Config
			result := &Config{}
			for _, key := range tc.keys {
				input := maximalConfig(key)
				inputs = append(inputs, input)
				result.Patch(input)
			}
			for _, input := range inputs {
				requireNothingShared(t, result, input)
			}

			changeEverything(result)

			for i, input := range inputs {
				require.Equal(t, maximalConfig(tc.keys[i]), input)
			}
		})
	}
}

// requireNothingShared fails if a pointer, map or slice of result is also reachable from the input.
// The back pointer of a channel to its config and the config a reference was resolved from
// (dynamic.Config) are not compared.
func requireNothingShared(t *testing.T, result *Config, input *Config) {
	t.Helper()
	inputMemory := map[memory]bool{}
	walkMemory(reflect.ValueOf(input).Elem(), map[memory]bool{}, func(m memory) { inputMemory[m] = true })
	walkMemory(reflect.ValueOf(result).Elem(), map[memory]bool{}, func(m memory) {
		require.False(t, inputMemory[m], "result shares %v at %x with a patch config", m.typ, m.addr)
	})
}

type memory struct {
	addr uintptr
	typ  reflect.Type
}

func isBackPointer(t reflect.Type) bool {
	return t == reflect.TypeOf(&dynamic.Config{}) || t == reflect.TypeOf(&Config{})
}

func walkMemory(v reflect.Value, seen map[memory]bool, visit func(memory)) {
	var m memory
	switch v.Kind() {
	case reflect.Ptr, reflect.Map:
		if v.IsNil() || isBackPointer(v.Type()) {
			return
		}
		m = memory{v.Pointer(), v.Type()}
	case reflect.Slice:
		if v.Cap() == 0 {
			return
		}
		m = memory{v.Pointer(), v.Type()}
	case reflect.Interface:
		if !v.IsNil() {
			walkMemory(v.Elem(), seen, visit)
		}
		return
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
		if v.IsNil() || seen[v.Pointer()] || isBackPointer(v.Type()) {
			return
		}
		seen[v.Pointer()] = true
		change(v.Elem(), seen)
	case reflect.Interface:
		// the schema of a SchemaRef is a pointer in an interface, other values are taken over
		if !v.IsNil() && v.Elem().Kind() == reflect.Ptr {
			change(v.Elem(), seen)
		}
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
