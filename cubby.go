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

// SetToExpire adds or updates the item value with an expiration date after
// interval mapped to key in the cache.
func (c *Cache[T]) SetToExpire(key string, value T, interval time.Duration) {
	now := time.Now().UTC()
	c.SetItem(key, Item[T]{
		Value:     value,
		CreatedAt: now,
		ExpiredAt: now.Add(interval),
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

// TimedCache extends Cache with functionality to routinely delete (sweep)
// expired items from the cache after a specified interval.
type TimedCache[T any] struct {
	*Cache[T]
	ticker *time.Ticker
}

// StartSweeping clears expired items from the cache at every interval.
func (tc *TimedCache[T]) StartSweep(interval time.Duration) {
	tc.ticker = time.NewTicker(interval)
	for range tc.ticker.C {
		tc.ClearExpired()
	}
}

// StopSweep stops the started sweeping routine.
func (tc *TimedCache[T]) StopSweep() {
	if tc.ticker != nil {
		tc.ticker.Stop()
	}
}

// NewTimedCache creates a TimedCache with values of a specified type that
// clears expired items at every interval.
func NewTimedCache[T any](interval time.Duration) *TimedCache[T] {
	cache := NewCache[T]()
	timedCache := &TimedCache[T]{Cache: cache}
	go timedCache.StartSweep(interval)
	return timedCache
}
