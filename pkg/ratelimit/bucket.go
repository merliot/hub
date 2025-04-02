// Package ratelimit provides a simple token bucket implementation for rate
// limiting.
package ratelimit

import (
	"sync"
	"time"
)

// bucket represents a token bucket that fills at a specified rate.
type bucket struct {
	mu           sync.Mutex    // Ensures thread-safe access to the bucket's state
	fillInterval time.Duration // Time between adding each token
	capacity     int64         // Maximum number of tokens the bucket can hold
	available    int64         // Current number of available tokens
	lastUpdate   time.Time     // Last time the bucket was updated
}

// newBucket creates a new bucket with the specified fill interval and capacity.
// The bucket is initialized with the maximum number of tokens.
func newBucket(fillInterval time.Duration, capacity int64) *bucket {
	return &bucket{
		fillInterval: fillInterval,
		capacity:     capacity,
		available:    capacity,
		lastUpdate:   time.Now(),
	}
}

// take takes up to the requested number of tokens from the bucket and
// returns the number of tokens actually taken. If count is less than or equal
// to zero, it returns 0 without modifying the bucket.
func (b *bucket) take(count int64) int64 {
	if count <= 0 {
		return 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Calculate tokens to add based on elapsed time since last update
	now := time.Now()
	elapsed := now.Sub(b.lastUpdate)
	tokensToAdd := int64(elapsed / b.fillInterval)

	// Update available tokens, capping at capacity
	b.available += tokensToAdd
	if b.available > b.capacity {
		b.available = b.capacity
	}

	// Update the last update time based on tokens added
	b.lastUpdate = b.lastUpdate.Add(time.Duration(tokensToAdd) * b.fillInterval)

	// Take up to the requested number of tokens
	taken := count
	if taken > b.available {
		taken = b.available
	}
	b.available -= taken

	return taken
}
