// Package cache provides a simple ttl-based item cache
package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache provides a simple item store
type Cache interface {
	// Insert the string k
	Insert(k string)
	// Exists proves the existence or lack-thereof of the item k in the cache
	Exists(k string) (exists bool)
}

type cache struct {
	u *gocache.Cache
}

func (c *cache) Insert(k string) {
	c.u.SetDefault(k, struct{}{})
}

func (c *cache) Exists(k string) (exists bool) {
	_, exists = c.u.Get(k)
	return exists
}

// NewCacheWithTTL creates a new cache with ttl of t
func NewCacheWithTTL(t time.Duration) Cache {
	return &cache{
		u: gocache.New(t, time.Minute*60),
	}
}
