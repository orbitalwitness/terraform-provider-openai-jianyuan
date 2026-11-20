package provider

import (
	"context"
	"sync"
)

// listCaches holds one listCache per configured client, so each provider instance (one plan or apply) refreshes
// high-cardinality resources with one paginated list per parent rather than one GET per resource.
var listCaches sync.Map

type listCache struct {
	mu      sync.Mutex
	entries map[string]*listCacheEntry
}

type listCacheEntry struct {
	mu    sync.Mutex
	items any
}

func (r *baseResource) listCache() *listCache {
	cache, _ := listCaches.LoadOrStore(r.client, &listCache{})
	return cache.(*listCache)
}

func (c *listCache) entry(scope string) *listCacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = map[string]*listCacheEntry{}
	}
	if c.entries[scope] == nil {
		c.entries[scope] = &listCacheEntry{}
	}
	return c.entries[scope]
}

func (c *listCache) invalidate(scope string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, scope)
}

// cachedGet returns the item with the given ID from the list for scope, calling list at most once per scope until
// it succeeds or the scope is invalidated. A nil item means the ID is absent from the list.
func cachedGet[T any](ctx context.Context, c *listCache, scope, id string, list func(context.Context) (map[string]T, error)) (*T, error) {
	entry := c.entry(scope)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.items == nil {
		items, err := list(ctx)
		if err != nil {
			return nil, err
		}
		entry.items = items
	}
	item, ok := entry.items.(map[string]T)[id]
	if !ok {
		return nil, nil
	}
	return &item, nil
}
