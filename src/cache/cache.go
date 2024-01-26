package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	m      *sync.Map
	ttls   map[K]time.Time
	ttlsMu *sync.Mutex
}

func New[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		m:      &sync.Map{},
		ttls:   map[K]time.Time{},
		ttlsMu: &sync.Mutex{},
	}
}

func (c Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.m.Load(key)
	var val V
	if v != nil {
		val = v.(V)
	}
	return val, ok
}

func (c Cache[K, V]) Set(key K, val V, dur time.Duration) {
	c.m.Swap(key, val)
	if dur == 0 {
		return
	}

	go func() {
		c.ttlsMu.Lock()
		c.ttls[key] = time.Now().Add(dur)
		c.ttlsMu.Unlock()

		time.Sleep(dur)

		c.ttlsMu.Lock()
		defer c.ttlsMu.Unlock()

		currTTL := c.ttls[key]
		if currTTL != (time.Time{}) && time.Now().After(currTTL) {
			c.Delete(key)
			c.ttls[key] = time.Time{}
		}
	}()
}

func (c Cache[K, V]) Delete(key K) {
	c.m.Delete(key)
}
