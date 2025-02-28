package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goudeaux.com/img"
)

func TestCache(t *testing.T) {
	c := New(2)
	err := c.Set("key1", "val1", time.Minute)
	assert.NoError(t, err)
	err = c.Set("key2", "val2", 2*time.Minute)
	assert.NoError(t, err)
	err = c.Set("key3", "val3", 3*time.Minute)
	assert.ErrorIs(t, err, ErrCapacity)

	val, err := c.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "val1", val)

	err = c.Set("key1", "newVal1", time.Minute)
	assert.NoError(t, err)

	val, err = c.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "newVal1", val)

	err = c.Delete("key2")
	assert.NoError(t, err)

	_, err = c.Get("key2")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = c.Get("unknownKey")
	assert.ErrorIs(t, err, ErrNotFound)

	err = c.Delete("unknownKey")
	assert.ErrorIs(t, err, ErrNotFound)

	err = c.Set("key4", "val4", 4*time.Minute)
	assert.NoError(t, err)

	stats := c.Stats()
	expected := img.Stats{
		Hits:      2,
		Misses:    2,
		Evictions: 1,
	}
	assert.Equal(t, expected, stats)

	c.Delete("key1")
	c.Delete("key4")

	c.Prune()
	stats = c.Stats()
	expected = img.Stats{
		Hits:      2,
		Misses:    2,
		Evictions: 3,
	}
	assert.Equal(t, expected, stats)
}
