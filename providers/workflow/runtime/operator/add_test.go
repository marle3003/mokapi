package operator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddInt(t *testing.T) {
	out, err := Add(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, out)
}

func TestAddFloat(t *testing.T) {
	out, err := Add(1.1, 2.2)
	assert.NoError(t, err)
	assert.Equal(t, 3.3, out)
}

func TestAddFloatConvert(t *testing.T) {
	out, err := Add(1, 2.2)
	assert.NoError(t, err)
	assert.Equal(t, 3.2, out)
}

func TestAddFloatConvert2(t *testing.T) {
	out, err := Add(1.1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, out)
}

func TestAddString(t *testing.T) {
	out, err := Add("a", "b")
	assert.NoError(t, err)
	assert.Equal(t, "ab", out)
}

func TestAddStringInt(t *testing.T) {
	out, err := Add("a", 1)
	assert.NoError(t, err)
	assert.Equal(t, "a1", out)
}

func TestAddStringFloat(t *testing.T) {
	out, err := Add("a", 1.1)
	assert.NoError(t, err)
	assert.Equal(t, "a1.1", out)
}
