package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Configuration
type Config struct {
	// Time between adding each token to a bucket
	FillInterval time.Duration
	// Maximum number of tokens in a bucket
	Capacity int64
	// This determines how often the rate limiter cleans up expired entries
	// (rate limiters for IP addresses that haven't been seen for a while).
	CleanupInterval time.Duration
}

type client struct {
	*bucket
	lastSeen time.Time
}

type RateLimiter struct {
	Config
	sync.Map // key: client IP address, value: *client
}

// New creates a new RateLimiter instance.
func New(config Config) *RateLimiter {
	rl := &RateLimiter{
		Config: config,
	}
	go rl.cleanupRateLimiters()
	return rl
}

// rateLimit returns an HTTP middleware that performs rate limiting.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddress(r)

		c, ok := rl.Load(ip)
		if !ok {
			c = &client{
				bucket:   newBucket(rl.FillInterval, rl.Capacity),
				lastSeen: time.Now(),
			}
			rl.Store(ip, c)
		}
		client := c.(*client)

		if client.bucket.take(1) == 0 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		client.lastSeen = time.Now()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) Stats() map[string]int64 {
	stats := make(map[string]int64)
	rl.Range(func(key, value any) bool {
		ip := key.(string)
		client := value.(*client)
		stats[ip] = client.bucket.avail()
		return true
	})
	return stats
}

func (rl *RateLimiter) cleanupRateLimiters() {

	ticker := time.NewTicker(rl.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.Range(func(key, value any) bool {
			ip := key.(string)
			client := value.(*client)
			if time.Since(client.lastSeen) > rl.CleanupInterval {
				rl.Delete(ip)
			}

			return true
		})
	}
}

func getIPAddress(r *http.Request) string {
	// 1. Check X-Forwarded-For header (important for proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, separated by commas.
		// We take the first one (the client's original IP).
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// 2. Fallback to RemoteAddr (less reliable, but often necessary)
	ip, _, err := net.SplitHostPort(r.RemoteAddr) // net.SplitHostPort handles IPv6 correctly
	if err != nil {
		return "0.0.0.0"
	}
	return ip
}
