# cubby ![tests](https://github.com/novrin/cubby/workflows/tests/badge.svg)

`cubby` is a tiny Go (golang) library for simple, type-safe, thread-safe, in-memory caches.

Use the provided `Cache` struct to store items of any specified type in a `map[string]Item[T]`. Optionally set these items to expire after a specified lifetime.

## Features

* **Tiny** - less than 200 LOC
* **Flexible** - intialize a cache to store any single type
* **Type-safe** - ensure all items in the same cache are of the same type
* **Thread-safe** - avoid unintended effects during concurrent access
* **In-memory** - eliminate the need to send data over a network


## Installation

```shell
go get github.com/novrin/cubby
``` 

## Usage
**Cache**

The `Cache` struct can be used to store items of ANY specified type.

```go
package main

import (
	"fmt"
	"time"

    "github.com/novrin/cubby"
)

func main() {
    // Use NewCache to instantiate a cache of ANY type. 
    // To illustrate, we'll use int.
    cache := cubby.NewCache[int]()

    // Use Set to map strings to item values.
	cache.Set("foo", 7)
	cache.Set("bar", 8)

	// Use SetToExpire to map strings to item values that expire.
	cache.SetToExpire("baz", 9, 5*time.Minute)

	// Use Get to retrieve item values mapped to the given key.
	foo, ok := cache.Get("foo")
	if ok {
		fmt.Println(foo)
	}

	// Use Delete to remove items mapped to the given key.
	cache.Delete("bar")

	// Use ClearExpired to remove all expired items.
	cache.ClearExpired()

	// Use Clear to remove all items.
	cache.Clear()
}
```
**TickingCache**

If you want a cache to run a function in timed intervals, you can use the provided `TickingCache` struct and assign its `Job` function. 

A common use case is to clear expired items at every interval:

```go
// Use NewTickingCache to instantiate a cache that runs its Job function
// at every duration - in this case, 3 hours.
cache := NewTickingCache[float32](3 * time.Hour)

// Assign the cache Job function. Here, we clear the cache.
cache.Job = func() {
    fmt.Println("Clearing the cache!")
    cache.ClearExpired()
}

keys := []string{"foo", "bar", "baz"}
values := []float32{3.14, 1.618, 2.718}
for i, k := range keys {
    cache.SetToExpire(k, values[i], 5*time.Minute)
}
// After 5 minutes, the items above will have expired but will remain in
// the cache for another 2hrs 55 minutes (until the very first tick runs
// its assigned job to clear the cache).
```