package cache

import "sync"

type Cache struct {
	lock  sync.RWMutex
	store map[string]interface{}
}

func NewPool() *Cache {
	return &Cache{store: make(map[string]interface{})}
}

func (c *Cache) Set(k string, v interface{}) {
	c.lock.Lock()
	c.store[k] = v
	c.lock.Unlock()
}

func (c *Cache) Get(k string) (interface{}, bool) {
	c.lock.RLock()
	v, ok := c.store[k]
	c.lock.RUnlock()
	return v, ok
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
