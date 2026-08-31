package portal

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const defaultEntitlementsTTL = 5 * time.Minute

type cachedEntry struct {
	value      *Entitlements
	fetchedAt  time.Time
	refreshing bool
}

func (e *cachedEntry) stale(ttl time.Duration) bool {
	return time.Since(e.fetchedAt) > ttl
}

// EntitlementsCache is a per-user stale-while-revalidate cache for Okta entitlements.
// The first request for a user is synchronous (one Okta call). Subsequent requests return
// the cached value immediately; a background refresh starts when the entry ages past the TTL.
// Concurrent cold-start requests for the same user are collapsed by singleflight so Okta
// never sees a burst of identical calls.
type EntitlementsCache struct {
	mu      sync.Mutex
	entries map[string]*cachedEntry
	ttl     time.Duration
	group   singleflight.Group
	fetch   func(ctx context.Context, sub string) (*Entitlements, error)
}

func newEntitlementsCache(ttl time.Duration, fetch func(ctx context.Context, sub string) (*Entitlements, error)) *EntitlementsCache {
	if ttl <= 0 {
		ttl = defaultEntitlementsTTL
	}
	return &EntitlementsCache{
		entries: make(map[string]*cachedEntry),
		ttl:     ttl,
		fetch:   fetch,
	}
}

// Get returns entitlements for the given user. On a cache miss it fetches synchronously.
// On a stale hit it returns the cached value immediately and refreshes in the background.
func (c *EntitlementsCache) Get(ctx context.Context, sub string) (*Entitlements, error) {
	c.mu.Lock()
	entry := c.entries[sub]
	c.mu.Unlock()

	if entry != nil && !entry.stale(c.ttl) {
		return entry.value, nil
	}

	if entry != nil {
		c.tryRefresh(sub)
		return entry.value, nil
	}

	v, err, _ := c.group.Do(sub, func() (interface{}, error) {
		ent, err := c.fetch(ctx, sub)
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		c.entries[sub] = &cachedEntry{value: ent, fetchedAt: time.Now()}
		c.mu.Unlock()
		return ent, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Entitlements), nil
}

// tryRefresh starts a background goroutine to refresh the entry for sub. It is a no-op if a
// refresh is already in flight for this user.
func (c *EntitlementsCache) tryRefresh(sub string) {
	c.mu.Lock()
	entry := c.entries[sub]
	if entry == nil || entry.refreshing {
		c.mu.Unlock()
		return
	}
	entry.refreshing = true
	c.mu.Unlock()

	go func() {
		ent, err := c.fetch(context.Background(), sub)
		c.mu.Lock()
		defer c.mu.Unlock()
		if err != nil {
			slog.Warn("background entitlements refresh failed", "sub", sub, "error", err)
			if c.entries[sub] != nil {
				c.entries[sub].refreshing = false
			}
			return
		}
		c.entries[sub] = &cachedEntry{value: ent, fetchedAt: time.Now()}
	}()
}
