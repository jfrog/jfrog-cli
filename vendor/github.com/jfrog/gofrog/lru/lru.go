package lru

import (
	"sync"
	"time"
)

type Cache struct {
	cache  *cacheBase
	lock   sync.Mutex
	noSync bool
}

func New(size int, options ...func(*Cache)) *Cache {
	c := &Cache{cache: newCacheBase(size)}
	for _, option := range options {
		option(c)
	}
	return c
}

func WithExpiry(expiry time.Duration) func(c *Cache) {
	return func(c *Cache) {
		c.cache.Expiry = expiry
	}
}

func WithEvictionCallback(onEvicted func(key string, value interface{})) func(c *Cache) {
	return func(c *Cache) {
		c.cache.OnEvicted = onEvicted
	}
}

func WithoutSync() func(c *Cache) {
	return func(c *Cache) {
		c.noSync = true
	}
}

func (c *Cache) Add(key string, value interface{}) {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	c.cache.Add(key, value)
}

func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	return c.cache.Get(key)
}

// Updates element's value without updating it's "Least-Recently-Used" status
func (c *Cache) UpdateElement(key string, value interface{}) {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	c.cache.UpdateElement(key, value)

}

func (c *Cache) Remove(key string) {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	c.cache.Remove(key)
}

func (c *Cache) RemoveOldest() {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	c.cache.RemoveOldest()
}

func (c *Cache) Len() int {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	return c.cache.Len()
}

func (c *Cache) Clear() {
	if !c.noSync {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	c.cache.Clear()
}
