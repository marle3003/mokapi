package cli

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchField(t *testing.T) {
	type embedded struct{ Inner string }
	type config struct {
		Plain  string
		Tagged string   `name:"custom-name"`
		Alias  string   `aliases:"first-alias,second"`
		Flag   string   `flag:"by-flag" name:"ignored"`
		Items  []string `explode:"item"`
		embedded
	}

	testcases := []struct {
		name     string
		paths    []string
		field    string
		consumed int
	}{
		{name: "field name is case insensitive", paths: []string{"plain", "more"}, field: "Plain", consumed: 1},
		{name: "name tag over several path segments", paths: []string{"custom", "name", "more"}, field: "Tagged", consumed: 2},
		{name: "aliases tag", paths: []string{"first", "alias"}, field: "Alias", consumed: 2},
		{name: "second alias", paths: []string{"second"}, field: "Alias", consumed: 1},
		{name: "flag tag wins over name tag", paths: []string{"by", "flag"}, field: "Flag", consumed: 2},
		{name: "promoted field of embedded struct", paths: []string{"inner"}, field: "Inner", consumed: 1},
		{name: "no match", paths: []string{"unknown"}, field: ""},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			index, consumed, ok := matchField(reflect.TypeOf(config{}), tc.paths)

			if tc.field == "" {
				require.False(t, ok)
				return
			}
			require.True(t, ok)
			require.Equal(t, tc.field, reflect.TypeOf(config{}).FieldByIndex(index).Name)
			require.Equal(t, tc.consumed, consumed)
		})
	}

	t.Run("explode tag", func(t *testing.T) {
		require.Equal(t, 4, fieldIndexByTag(reflect.TypeOf(config{}), "item", "explode"))
		require.Equal(t, -1, fieldIndexByTag(reflect.TypeOf(config{}), "other", "explode"))
	})
}
