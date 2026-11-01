package try

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func VerifySubset(t *testing.T, main any, sub map[string]any) {
	v := reflect.ValueOf(main)
	if v.Kind() != reflect.Map {
		t.Errorf("main has wrong kind of value %t", main)
	}

	for k, wantVal := range sub {
		mv := v.MapIndex(reflect.ValueOf(k))
		if !mv.IsValid() {
			t.Errorf("missing key %q", k)
			continue
		}
		gotVal := mv.Interface()
		if m, ok := gotVal.(map[string]any); ok {
			subWantVal, isMap := wantVal.(map[string]any)
			if !isMap {
				t.Errorf("key %q has wrong type %T", k, wantVal)
			}
			VerifySubset(t, m, subWantVal)
		} else {
			require.Equal(t, wantVal, gotVal)
		}
	}
}
