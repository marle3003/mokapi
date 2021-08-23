package lua

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestString(t *testing.T) {
	s := Dump("foo")
	assert.Equal(t, "foo", s)
}

func TestStruct(t *testing.T) {
	type T struct {
		A string
		B int
	}

	s := Dump(&T{A: "foo", B: 42})
	assert.Equal(t, "T{A: foo, B: 42}", s)
}

func TestNestedStruct(t *testing.T) {
	type T struct {
		A string
		B *T
	}

	s := Dump(&T{A: "foo", B: &T{A: "bar"}})
	assert.Equal(t, "T{A: foo, B: T{A: bar, B: <nil>}}", s)
}
