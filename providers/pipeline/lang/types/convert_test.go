package types

import (
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	data := []struct {
		test string
		in   interface{}
		out  Object
	}{
		{"int", 1, NewNumber(1.0)},
		{"float", 1.2, NewNumber(1.2)},
		{"string", "foobar", NewString("foobar")},
		{"bool", false, NewBool(false)},
		{"bool", true, NewBool(true)},
		{"expando", map[string]interface{}{"a": 1, "b": "foobar", "isValid": true}, func() Object {
			e := NewExpando()
			e.SetField("a", NewNumber(1))
			e.SetField("b", NewString("foobar"))
			e.SetField("isValid", NewBool(true))
			return e
		}()},
		{"ref", struct{ foo string }{"bar"}, NewReference(struct{ foo string }{"bar"})},
	}

	for _, d := range data {
		o, err := Convert(d.in)
		if err != nil {
			t.Errorf("convert(%q):%v", d.test, err.Error())
		}
		if !reflect.DeepEqual(o, d.out) {
			t.Errorf("convert(%q): got %q, expected %q", d.test, o, d.out)
		}
	}
}

func TestConvertFrom(t *testing.T) {
	data := []struct {
		test string
		in   Object
		to   reflect.Type
		out  interface{}
	}{
		{"int", NewNumber(1.0), reflect.TypeOf(1), 1},
		{"int", NewNumber(1.3), reflect.TypeOf(1), 1},
		{"float", NewNumber(1.3), reflect.TypeOf(1.0), 1.3},
		{"string", NewString("foobar"), reflect.TypeOf(""), "foobar"},
	}

	for _, d := range data {
		o, err := ConvertFrom(d.in, d.to)
		if err != nil {
			t.Errorf("convert(%q):%v", d.test, err.Error())
		}
		if !reflect.DeepEqual(o, d.out) {
			t.Errorf("convert(%q): got %q, expected %q", d.test, o, d.out)
		}
	}
}
