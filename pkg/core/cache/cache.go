package cache

import "sync"

type Cache struct {
	store sync.Map
}

func NewCache() *Cache {
	var m Cache
	return &m
}

func (c *Cache) Set(k string, v interface{}) {
	c.store.Store(k, v)
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (c *Cache) GetOrSet(k string, v interface{}) (interface{}, bool) {
	return c.store.LoadOrStore(k, v)
}

func (c *Cache) Get(k string) (interface{}, bool) {
	return c.store.Load(k)
}

func (c *Cache) Range(f func(key, value interface{}) bool) {
	c.store.Range(f)
}

func (c *Cache) Clean() {
	c.store = sync.Map{}
}

func (c *Cache) GetMustInt(k string) (int, bool) {
	v, ok := c.Get(k)
	return v.(int), ok
}

func (c *Cache) GetMustString(k string) (string, bool) {
	v, ok := c.Get(k)
	return v.(string), ok
}

func (c *Cache) GetMustBool(k string) (bool, bool) {
	v, ok := c.Get(k)
	return v.(bool), ok
}
