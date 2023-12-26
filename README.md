# cubby

[![GoDoc](https://godoc.org/github.com/novrin/cubby?status.svg)](https://pkg.go.dev/github.com/novrin/cubby) 
![tests](https://github.com/novrin/cubby/workflows/tests/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/novrin/cubby)](https://goreportcard.com/report/github.com/novrin/cubby)

`cubby` is a tiny Go (golang) library for simple, type-safe, thread-safe, in-memory caches.

Store values of any type in keys of any comparable type and optionally set them to expire after a specified lifetime.

### Features

* **Tiny** - less than 150 LOC and no external dependencies
* **Flexible** - initialize caches with comparable keys and values of any type
* **Type-safe** - ensure a initialized cache uses consistent key and value types
* **Thread-safe** - avoid unintended effects during concurrent access
* **In-memory** - eliminate the need to send data over a network

### Installation

```shell
go get github.com/novrin/cubby
``` 

## Usage

### Cache

A `Cache` can be used to map keys of ANY comparable type to values of ANY type.

```go
package main

import (
	"fmt"
	"time"

    "github.com/novrin/cubby"
)

func main() {
	cache := cubby.NewCache[string, int]()

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

A `TickingCache` extends `Cache` with a ticker. In a single, new go routine, it runs an assigned `Job` function at every tick.

A common use case is to clear expired items in timed intervals:

```go
cache := cubby.NewTickingCache[string, float32](3 * time.Hour)

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

## License

Copyright (c) 2023-present [novrin](https://github.com/novrin)

Licensed under [MIT License](./LICENSE)