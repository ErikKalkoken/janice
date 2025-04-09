package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddToListWithRotation(t *testing.T) {
	t.Run("can add to empty list", func(t *testing.T) {
		var l []string
		l2 := addToListWithRotation(l, "alpha", 5)
		assert.Equal(t, []string{"alpha"}, l2)
	})
	t.Run("should insert new items on top", func(t *testing.T) {
		var l = []string{"alpha"}
		l2 := addToListWithRotation(l, "bravo", 5)
		assert.Equal(t, []string{"bravo", "alpha"}, l2)
	})
	t.Run("should throw away bottom item to keep max length", func(t *testing.T) {
		var l = []string{"alpha", "bravo", "charlie"}
		l2 := addToListWithRotation(l, "delta", 3)
		assert.Equal(t, []string{"delta", "alpha", "bravo"}, l2)
	})
	t.Run("should insert new on top and remove duplicates", func(t *testing.T) {
		var l = []string{"alpha", "bravo"}
		l2 := addToListWithRotation(l, "bravo", 5)
		assert.Equal(t, []string{"bravo", "alpha"}, l2)
	})
}
