package cubby

import (
	"sync"
	"time"
)

// Item represents a unit mapped to a key in a Cache.
type Item[V any] struct {
	Value     V
	CreatedAt time.Time
	ExpiredAt time.Time
}

// IsExpired returns true if time now is past the item's set ExpiredAt date.
func (i *Item[V]) IsExpired() bool {
	return !i.ExpiredAt.IsZero() && time.Now().UTC().After(i.ExpiredAt)
}

// Cache represents a generic store that wraps a map of a comparable type to
// an Item with a value of any type and a mutex for concurrent access.
type Cache[K comparable, V any] struct {
	items map[K]Item[V]
	mu    sync.RWMutex
}

// SetItem adds or updates the item mapped to key in the cache.
func (c *Cache[K, V]) SetItem(key K, item Item[V]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item
}

// Set adds or updates the item value mapped to key in the cache. CreatedAt is
// always set to time now.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetItem(key, Item[V]{
		Value:     value,
		CreatedAt: time.Now().UTC(),
	})
}

// SetToExpire adds or updates the item value with an expiration date equal to
// time now + lifetime mapped to key in the cache.
func (c *Cache[K, V]) SetToExpire(key K, value V, lifetime time.Duration) {
	now := time.Now().UTC()
	c.SetItem(key, Item[V]{
		Value:     value,
		CreatedAt: now,
		ExpiredAt: now.Add(lifetime),
	})
}

// GetItem retrieves the item mapped to key from the cache.
func (c *Cache[K, V]) GetItem(key K) (Item[V], bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	return item, ok
}

// Get retrieves the item value mapped to key from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	item, ok := c.GetItem(key)
	return item.Value, ok
}

// Delete removes the item mapped to key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]Item[V])
}

// ClearExpired removes all expired items from the cache.
func (c *Cache[K, V]) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
		}
	}
}

// Items returns a copy of the items map.
func (c *Cache[K, V]) Items() map[K]Item[V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[K]Item[V], len(c.items))
	for k, v := range c.items {
		items[k] = v
	}
	return items
}

// Len returns the length of the items map in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// NewCache creates a Cache with K type keys and V type values.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]Item[V]),
	}
}

// TickingCache extends Cache with functionality to process a job at every
// interval. A common application is to clear expired entries at every tick.
type TickingCache[K comparable, V any] struct {
	*Cache[K, V]
	ticker *time.Ticker
	Job    func()
}

// Start creates a new ticker and calls Job at every tick denoted by duration.
func (tc *TickingCache[k, V]) Start(d time.Duration) {
	tc.ticker = time.NewTicker(d)
	for range tc.ticker.C {
		if tc.Job != nil {
			tc.Job()
		}
	}
}

// Stop immediately stops ticking to prevent Job from being called.
func (tc *TickingCache[K, V]) Stop() {
	if tc.ticker != nil {
		tc.ticker.Stop()
	}
}

// NewTickingCache creates a Cache with K type keys and V type values and starts
// a single, new go routine that calls job at every tick denoted by duration.
func NewTickingCache[K comparable, V any](d time.Duration) *TickingCache[K, V] {
	tc := &TickingCache[K, V]{Cache: NewCache[K, V]()}
	go tc.Start(d)
	return tc
}
