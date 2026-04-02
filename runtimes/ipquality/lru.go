package ipquality

import lru "github.com/hashicorp/golang-lru"

type MemoryCache struct {
	lru *lru.Cache
}

func NewMemoryCache(size int) *MemoryCache {
	c, _ := lru.New(size)
	return &MemoryCache{lru: c}
}

func (c *MemoryCache) Get(ip string) (*Decision, bool) {
	if v, ok := c.lru.Get(ip); ok {
		return v.(*Decision), true
	}
	return nil, false
}

func (c *MemoryCache) Set(ip string, d *Decision) {
	c.lru.Add(ip, d)
}
