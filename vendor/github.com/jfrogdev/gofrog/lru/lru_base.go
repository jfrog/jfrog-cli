// Package go-lru implements an LRU cache.
// It is based on the
// LRU implementation in groupcache:
// https://github.com/golang/groupcache/tree/master/lru
package lru

import (
	"container/list"
	"time"
)

type cacheBase struct {
	Expiry time.Duration
	Size   int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key string, value interface{})

	ll    *list.List
	cache map[string]*list.Element
}

type entry struct {
	key        string
	value      interface{}
	timeInsert int64
}

func newCacheBase(size int) *cacheBase {
	return &cacheBase{Size: size, cache: make(map[string]*list.Element), ll: list.New()}
}

func (c *cacheBase) Add(key string, value interface{}) {
	var epochNow int64
	if c.Expiry != time.Duration(0) {
		epochNow = time.Now().UnixNano() / int64(time.Millisecond)
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		ee.Value.(*entry).timeInsert = epochNow
		return
	}
	ele := c.ll.PushFront(&entry{key, value, epochNow})
	c.cache[key] = ele
	if c.Size != 0 && c.ll.Len() > c.Size {
		c.RemoveOldest()
	}
	return
}

func (c *cacheBase) Get(key string) (value interface{}, ok bool) {
	if ele, hit := c.cache[key]; hit {
		if c.Expiry != time.Duration(0) {
			unixNow := time.Now().UnixNano() / int64(time.Millisecond)
			unixExpiry := int64(c.Expiry / time.Millisecond)
			if (unixNow - ele.Value.(*entry).timeInsert) > unixExpiry {
				c.removeElement(ele)
				return nil, false
			}
		}
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

// Updates element's value without updating it's "Least-Recently-Used" status
func (c *cacheBase) UpdateElement(key string, value interface{}) {
	if ee, ok := c.cache[key]; ok {
		ee.Value.(*entry).value = value
		return
	}
}

func (c *cacheBase) Remove(key string) {
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *cacheBase) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *cacheBase) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *cacheBase) Len() int {
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *cacheBase) Clear() {
	for _, e := range c.cache {
		kv := e.Value.(*entry)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
		delete(c.cache, kv.key)
	}
	c.ll.Init()
}
