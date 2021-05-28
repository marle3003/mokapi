package operator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubstractInt(t *testing.T) {
	out, err := Substract(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, -1, out)
}

func TestSubstractFloat(t *testing.T) {
	out, err := Substract(2.2, 1.1)
	assert.NoError(t, err)
	assert.Equal(t, 1.1, out)
}

func TestSubstractFloatConvert(t *testing.T) {
	out, err := Substract(5, 2.2)
	assert.NoError(t, err)
	assert.Equal(t, 2.8, out)
}

func TestSubstractFloatConvert2(t *testing.T) {
	out, err := Substract(2.1, 2)
	assert.NoError(t, err)
	assert.InDelta(t, 0.1, out, 0.00000001)
}
