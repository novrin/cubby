package cubby

import (
	"sync"
	"time"
)

// Item represents a unit storable in a Cache.
type Item[T any] struct {
	Value     T
	CreatedAt time.Time
	ExpiredAt time.Time
}

// IsExpired returns true if the current time is after the item's explicity set expiration.
func (e *Item[T]) IsExpired() bool {
	return !e.ExpiredAt.IsZero() && time.Now().UTC().After(e.ExpiredAt)
}

// Cache represents a generic store for a specified type with a map and mutex
// for concurrent access.
type Cache[T any] struct {
	items map[string]Item[T]
	mu    sync.Mutex
}

// SetItem adds or updates the item mapped to key in the cache.
func (c *Cache[T]) SetItem(key string, item Item[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item
}

// Set adds or updates the item value mapped to key in the cache. CreatedAt is
// always set to the current time.
func (c *Cache[T]) Set(key string, value T) {
	c.SetItem(key, Item[T]{
		Value:     value,
		CreatedAt: time.Now().UTC(),
	})
}

// SetToExpire adds or updates the item value with an expiration date equal to
// the current time + lifetime mapped to key in the cache.
func (c *Cache[T]) SetToExpire(key string, value T, lifetime time.Duration) {
	now := time.Now().UTC()
	c.SetItem(key, Item[T]{
		Value:     value,
		CreatedAt: now,
		ExpiredAt: now.Add(lifetime),
	})
}

// GetItem retrieves the item mapped to key from the cache.
func (c *Cache[T]) GetItem(key string) (Item[T], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	return item, ok
}

// Get retrieves the item value mapped to key from the cache.
func (c *Cache[T]) Get(key string) (T, bool) {
	item, ok := c.GetItem(key)
	return item.Value, ok
}

// Delete removes the item mapped to key from the cache.
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache.
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item[T])
}

// ClearExpired removes all expired items from the cache.
func (c *Cache[T]) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
		}
	}
}

// Items returns a copy of the items map.
func (c *Cache[T]) Items() map[string]Item[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	items := make(map[string]Item[T], len(c.items))
	for k, v := range c.items {
		items[k] = v
	}
	return items
}

// Len returns the length of the items map in the cache.
func (c *Cache[T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// NewCache creates a Cache with values of a specified type.
func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		items: make(map[string]Item[T]),
	}
}

// TickingCache extends Cache with functionality to process a job at every
// interval. A common application is to clear expired entries at every tick.
type TickingCache[T any] struct {
	*Cache[T]
	ticker *time.Ticker
	Job    func()
}

// Start creates a new ticker and calls job at every tick denoted by duration.
func (tc *TickingCache[T]) Start(d time.Duration) {
	tc.ticker = time.NewTicker(d)
	for range tc.ticker.C {
		if tc.Job != nil {
			tc.Job()
		}
	}
}

// Stop immediately stops ticking to prevent job from being called.
func (tc *TickingCache[T]) Stop() {
	if tc.ticker != nil {
		tc.ticker.Stop()
	}
}

// NewTickingCache creates a Cache with values of a specified type and starts
// a new go routine that calls job at every tick denoted by duration.
func NewTickingCache[T any](d time.Duration) *TickingCache[T] {
	cache := NewCache[T]()
	timedCache := &TickingCache[T]{Cache: cache}
	go timedCache.Start(d)
	return timedCache
}
