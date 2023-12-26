package cubby

import (
	"strconv"
	"testing"
	"time"
)

const errorString = "\nGot: %v\nWanted: %v\n"

func TestIsExpired(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	type unit struct {
		item Item[int]
		want bool
	}
	cases := map[string]unit{
		"no expiration": {
			item: Item[int]{Value: 1, CreatedAt: now},
			want: false,
		},
		"future expiration": {
			item: Item[int]{Value: 2, CreatedAt: now, ExpiredAt: future},
			want: false,
		},
		"past expiration": {
			item: Item[int]{Value: 3, CreatedAt: past, ExpiredAt: past},
			want: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.item.IsExpired()
			if got != tc.want {
				t.Fatalf(
					errorString,
					strconv.FormatBool(got),
					strconv.FormatBool(tc.want),
				)
			}
		})
	}
}

var keys = []string{"x", "y", "z"}

func TestCache(t *testing.T) {
	cache := NewCache[string, int]()
	for _, k := range keys {
		if _, ok := cache.Get(k); ok {
			t.Fatalf("Got value for %s but %s should not exist.", k, k)
		}
	}
}

func TestSetItem(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	cases := map[string]Item[int]{
		"now":    {Value: 1, CreatedAt: now},
		"future": {Value: 2, CreatedAt: now, ExpiredAt: future},
		"past":   {Value: 3, CreatedAt: past, ExpiredAt: past},
	}
	cache := NewCache[string, int]()
	for name, item := range cases {
		t.Run(name, func(t *testing.T) {
			cache.SetItem(name, item)
			want := cases[name]
			if got, ok := cache.GetItem(name); !ok || item != want {
				t.Fatalf(errorString, got, want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	cases := map[string][]int{
		"initial": {1, 2, 3},
		"updates": {7, 8, 9},
	}
	cache := NewCache[string, int]()
	for name, nums := range cases {
		t.Run(name, func(t *testing.T) {
			for i, k := range keys {
				v := nums[i]
				cache.Set(k, v)
				if x, ok := cache.Get(k); !ok || x != v {
					t.Fatalf(errorString, v, x)
				}
			}
		})
	}
}

func TestSetToExpire(t *testing.T) {
	cases := map[string]time.Duration{
		"second": 1 * time.Second,
		"minute": 1 * time.Minute,
		"hour":   1 * time.Hour,
		"day":    24 * time.Hour,
	}
	cache := NewCache[string, int]()
	for name, duration := range cases {
		t.Run(name, func(t *testing.T) {
			cache.SetToExpire(name, 1, duration)
			item, ok := cache.GetItem(name)
			if !ok {
				t.Fatalf("Wanted key %s to be in cache but it was not", name)
			}
			want := item.CreatedAt.Add(duration)
			if item.ExpiredAt != want {
				t.Fatalf(errorString, item.ExpiredAt, want)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cache := NewCache[string, int]()
	values := []int{1, 2, 3}
	for i, k := range keys {
		cache.Set(k, values[i])
		if _, ok := cache.Get(k); !ok {
			t.Fatalf("Wanted key %s to be in cache but it was not", k)
		}
		cache.Delete(k)
		if _, ok := cache.Get(k); ok { // item still in cache
			t.Fatalf("Wanted key %s to be deleted but it was not", k)
		}
	}
}

func TestClear(t *testing.T) {
	cache := NewCache[string, int]()
	values := []int{1, 2, 3}
	for i, k := range keys {
		cache.Set(k, values[i])
	}
	if cache.Len() != len(values) {
		t.Fatalf("Got cache length %v but wanted %v", cache.Len(), len(values))
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Fatalf("Got %v items but wanted cache to be empty", cache.Len())
	}
}

func TestClearExpired(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	type unit struct {
		items map[string]Item[int]
		want  int
	}
	cases := map[string]unit{
		"none expired": {
			items: map[string]Item[int]{
				"noEx1": {Value: 1, CreatedAt: now},
				"noEx2": {Value: 2, CreatedAt: now, ExpiredAt: future},
				"noEx3": {Value: 3, CreatedAt: past, ExpiredAt: future},
			},
			want: 3,
		},
		"one expired": {
			items: map[string]Item[int]{
				"noEx1": {Value: 1, CreatedAt: now},
				"noEx2": {Value: 2, CreatedAt: now, ExpiredAt: future},
				"ex":    {Value: 3, CreatedAt: past, ExpiredAt: past},
			},
			want: 2,
		},
		"all expired": {
			items: map[string]Item[int]{
				"ex1": {Value: 1, CreatedAt: past, ExpiredAt: past},
				"ex2": {Value: 2, CreatedAt: past, ExpiredAt: past},
				"ex3": {Value: 3, CreatedAt: past, ExpiredAt: past},
			},
			want: 0,
		},
	}
	cache := NewCache[string, int]()
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.items {
				cache.SetItem(k, v)
			}
			if cache.Len() != len(tc.items) {
				t.Fatalf(errorString, cache.Len(), len(tc.items))
			}
			cache.ClearExpired()
			if cache.Len() != tc.want {
				t.Fatalf(errorString, cache.Len(), tc.want)
			}
			cache.Clear()
		})
	}
}

func TestItems(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)
	type unit struct {
		items map[string]Item[int]
		want  int
	}
	cases := map[string]unit{
		"none expired": {
			items: map[string]Item[int]{
				"noEx1": {Value: 1, CreatedAt: now},
				"noEx2": {Value: 2, CreatedAt: now, ExpiredAt: future},
				"noEx3": {Value: 3, CreatedAt: past, ExpiredAt: future},
			},
			want: 3,
		},
		"one expired": {
			items: map[string]Item[int]{
				"noEx1": {Value: 1, CreatedAt: now},
				"noEx2": {Value: 2, CreatedAt: now, ExpiredAt: future},
				"ex":    {Value: 3, CreatedAt: past, ExpiredAt: past},
			},
			want: 2,
		},
		"all expired": {
			items: map[string]Item[int]{
				"ex1": {Value: 1, CreatedAt: past, ExpiredAt: past},
				"ex2": {Value: 2, CreatedAt: past, ExpiredAt: past},
				"ex3": {Value: 3, CreatedAt: past, ExpiredAt: past},
			},
			want: 0,
		},
	}
	cache := NewCache[string, int]()
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.items {
				cache.SetItem(k, v)
			}
			copy := cache.Items()
			if len(copy) != cache.Len() {
				t.Fatalf(errorString, len(copy), cache.Len())
			}
			for k, v1 := range copy {
				if v2, ok := cache.GetItem(k); !ok || v1 != v2 {
					t.Fatalf(errorString, v1, v2)
				}
			}
		})
	}
}

func TestTickingCache(t *testing.T) {
	cache := NewTickingCache[string, int](1 * time.Minute)
	for _, k := range keys {
		if _, ok := cache.Get(k); ok {
			t.Fatalf("Got value for %s but %s should not exist.", k, k)
		}
	}
}

func TestTickingCacheStartAndStop(t *testing.T) {
	cache := NewTickingCache[string, int](5 * time.Millisecond)
	cache.Job = func() {
		cache.ClearExpired()
	}
	values := []int{1, 2, 3}
	for i, k := range keys {
		cache.SetToExpire(k, values[i], 1*time.Millisecond)
	}
	if cache.Len() != len(values) {
		t.Fatalf("Got cache length %v but wanted %v", cache.Len(), len(values))
	}
	time.Sleep(10 * time.Millisecond)
	if cache.Len() != 0 {
		t.Fatalf("Got %v items but wanted cache to be empty", cache.Len())
	}
	for i, k := range keys {
		cache.SetToExpire(k, values[i], 1*time.Millisecond)
	}
	if cache.Len() != len(values) {
		t.Fatalf("Got cache length %v but wanted %v", cache.Len(), len(values))
	}
	cache.Stop()
	time.Sleep(10 * time.Millisecond)
	if cache.Len() == 0 { // ticker did not stop and items were cleared
		t.Fatalf("Got empty cache but wanted to have %v items", cache.Len())
	}
}
