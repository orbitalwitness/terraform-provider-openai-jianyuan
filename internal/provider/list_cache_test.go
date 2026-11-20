package provider

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCachedGet(t *testing.T) {
	var cache listCache
	var calls atomic.Int32
	list := func(context.Context) (map[string]string, error) {
		calls.Add(1)
		return map[string]string{"a": "alpha"}, nil
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			item, err := cachedGet(t.Context(), &cache, "scope", "a", list)
			if err != nil || item == nil || *item != "alpha" {
				t.Errorf("got %v, %v", item, err)
			}
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Errorf("expected one list call, got %d", calls.Load())
	}

	if item, err := cachedGet(t.Context(), &cache, "scope", "missing", list); err != nil || item != nil {
		t.Errorf("missing item must be nil without error, got %v, %v", item, err)
	}

	cache.invalidate("scope")
	if _, err := cachedGet(t.Context(), &cache, "scope", "a", list); err != nil || calls.Load() != 2 {
		t.Errorf("invalidate must force a new list call, got %d calls, err %v", calls.Load(), err)
	}
}
