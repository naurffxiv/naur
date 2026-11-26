package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointerHelpersReturnDistinctPointersToTheValue(t *testing.T) {
	a, b := String("x"), String("x")
	assert.Equal(t, "x", *a)
	assert.NotSame(t, a, b)

	assert.True(t, *Bool(true))
	assert.False(t, *Bool(false))
	assert.Equal(t, 42, *Int(42))
}
