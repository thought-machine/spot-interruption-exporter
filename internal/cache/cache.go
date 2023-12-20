// Package cache provides a simple ttl-based item cache
package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache provides a simple item store
type Cache interface {
	// Insert the string k
	Insert(k string, v string)
	// Exists proves the existence or lack-thereof of the item k in the cache
	Exists(k string) (exists bool)
	// Get returns the value of the key in the cache
	Get(k string) (value string, ok bool)
	// SetExpiration sets the expiration on the given item k without updating the value
	SetExpiration(k string, t time.Duration)
}

const NoExpiration = gocache.NoExpiration

type cache struct {
	u *gocache.Cache
}

func (c *cache) SetExpiration(k string, t time.Duration) {
	e, ok := c.u.Get(k)
	if !ok {
		return
	}
	c.u.Set(k, e, t)
}

func (c *cache) Insert(k string, v string) {
	c.u.SetDefault(k, v)
}

func (c *cache) Exists(k string) (exists bool) {
	_, exists = c.u.Get(k)
	return exists
}

func (c *cache) Get(k string) (value string, ok bool) {
	v, ok := c.u.Get(k)
	if !ok {
		return "", ok
	}
	value, ok = v.(string)
	return value, ok
}

// NewCacheWithTTL creates a new cache with ttl of t
func NewCacheWithTTL(t time.Duration) Cache {
	return &cache{
		u: gocache.New(t, time.Minute*60),
	}
}

func NewCacheWithTTLFrom(t time.Duration, f map[string]string) Cache {
	m := make(map[string]gocache.Item, len(f))
	for k, v := range f {
		m[k] = gocache.Item{
			Object:     v,
			Expiration: t.Nanoseconds(),
		}
	}
	return &cache{
		u: gocache.NewFrom(t, time.Minute*60, m),
	}
}
