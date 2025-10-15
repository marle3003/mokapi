package util_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/js/util"
	"testing"
)

func TestJsType(t *testing.T) {
	require.Equal(t, "Array", util.JsType([]string{}))
	require.Equal(t, "Integer", util.JsType(int64(1)))
	require.Equal(t, "Number", util.JsType(float64(1.1)))
	require.Equal(t, "Boolean", util.JsType(true))
	require.Equal(t, "Object", util.JsType(map[string]any{}))
	require.Equal(t, "String", util.JsType("123"))
	require.Equal(t, "Unknown", util.JsType(123))
}
