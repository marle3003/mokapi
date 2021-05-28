package operator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiplyInt(t *testing.T) {
	out, err := Multiply(4, 2)
	assert.NoError(t, err)
	assert.Equal(t, 8, out)
}

func TestMultiplyInt2(t *testing.T) {
	out, err := Multiply(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, out)
}

func TestMultiplyFloat(t *testing.T) {
	out, err := Multiply(2.2, 3.5)
	assert.NoError(t, err)
	assert.InDelta(t, 7.7, out, 0.00000001)
}

func TestMultiplyFloatConvert(t *testing.T) {
	out, err := Multiply(3, 1.5)
	assert.NoError(t, err)
	assert.Equal(t, 4.5, out)
}

func TestMultiplyFloatConvert2(t *testing.T) {
	out, err := Multiply(3.5, 2)
	assert.NoError(t, err)
	assert.Equal(t, 7.0, out)
}
