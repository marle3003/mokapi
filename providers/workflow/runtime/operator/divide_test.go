package operator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDivideInt(t *testing.T) {
	out, err := Divide(4, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, out)
}

func TestDivideInt2(t *testing.T) {
	out, err := Divide(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 0, out)
}

func TestDivideFloat(t *testing.T) {
	out, err := Divide(7.7, 3.5)
	assert.NoError(t, err)
	assert.Equal(t, 2.2, out)
}

func TestDivideFloatConvert(t *testing.T) {
	out, err := Divide(3, 1.5)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, out)
}

func TestDivideFloatConvert2(t *testing.T) {
	out, err := Divide(3.5, 2)
	assert.NoError(t, err)
	assert.Equal(t, 1.75, out)
}
