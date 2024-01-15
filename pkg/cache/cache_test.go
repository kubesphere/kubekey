package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	testCache := NewLocalCache("test")
	assert.Equal(t, "test", testCache.Name())

	// should not be able to get the key
	_, ok := testCache.Get("foo")
	assert.False(t, ok)

	// put a key
	testCache.Put("foo", "bar")

	// should be able to get the key
	v, ok := testCache.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", v)

	// release the key
	testCache.Release("foo")

	// should not be able to get the key
	_, ok = testCache.Get("foo")
	assert.False(t, ok)
}
