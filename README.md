# cubby ![tests](https://github.com/novrin/cubby/workflows/tests/badge.svg)

`cubby` is a tiny Go (golang) library for simple, type-safe, thread-safe, in-memory caches.

Use the provided `Cache` to store items of any specified type in a `map[string]Item[T]`. Optionally set these items to expire after a specified lifetime.

### Features

* **Tiny** - less than 200 LOC
* **Flexible** - intialize a cache to store any single type
* **Type-safe** - ensure all items in the same cache are of the same type
* **Thread-safe** - avoid unintended effects during concurrent access
* **In-memory** - eliminate the need to send data over a network


### Installation

```shell
go get github.com/novrin/cubby
``` 

## Usage

### Cache

A `Cache` can be used to store items of ANY specified type.

```go
package main

import (
	"fmt"
	"time"

    "github.com/novrin/cubby"
)

func main() {
    cache := cubby.NewCache[int]()

    // Map strings to item values.
	cache.Set("foo", 7)
	cache.Set("bar", 8)

	// Map strings to item values that expire.
	cache.SetToExpire("baz", 9, 5*time.Minute)

	// Retrieve mapped item values.
	foo, ok := cache.Get("foo")
	if ok {
		fmt.Println(foo)
	}

	// Remove items.
	cache.Delete("bar")

	// Remove all expired items.
	cache.ClearExpired()

	// Remove all items.
	cache.Clear()
}
```

### TickingCache

A `TickingCache` extends `Cache` with a ticker. It runs an assigned `Job` function in a separate go routine at every tick.

A common use case is to clear expired items in timed intervals:

```go
cache := NewTickingCache[float32](3 * time.Hour)

// Assign the cache Job function.
cache.Job = func() {
    fmt.Println("Clearing the cache!")
    cache.ClearExpired()
}

keys := []string{"foo", "bar", "baz"}
values := []float32{3.14, 1.618, 2.718}
for i, k := range keys {
    cache.SetToExpire(k, values[i], 5*time.Minute)
}
// After 5 minutes, the items above will expire but remain in the cache.
// They will be removed only after the very first tick (a total of 3hrs later).
```