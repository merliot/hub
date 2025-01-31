package ratelimit

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/juju/ratelimit"
)

// Configuration
type Config struct {
	// The time window within which the rate limit is enforced. It's how
	// often the "bucket" refills.
	RateLimitWindow time.Duration
	// The maximum number of requests allowed within the RateLimitWindow.
	MaxRequests int
	// This allows for a small burst of requests that exceed the average
	// rate limit. It defines the maximum number of requests that can be
	// made immediately even if the bucket is not full. Think of it as the
	// bucket's capacity.
	BurstSize int
	// This determines how often the rate limiter cleans up expired entries
	// (rate limiters for IP addresses that haven't been seen for a while).
	CleanupInterval time.Duration
}

type ipRateLimiter struct {
	limiter  *ratelimit.Bucket
	lastSeen time.Time
}

type RateLimiter struct {
	config         Config
	ipRateLimiters sync.Map
}

// New creates a new RateLimiter instance.
func New(config Config) *RateLimiter {
	rl := &RateLimiter{
		config: config,
	}
	go rl.cleanupRateLimiters()
	return rl
}

// rateLimit returns an HTTP middleware that performs rate limiting.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddress(r)

		limiter, loaded := rl.ipRateLimiters.Load(ip)

		if !loaded {
			limiter = &ipRateLimiter{
				limiter:  ratelimit.NewBucket(rl.config.RateLimitWindow, int64(rl.config.BurstSize)),
				lastSeen: time.Now(),
			}
			rl.ipRateLimiters.Store(ip, limiter)
		}
		ipLimiter := limiter.(*ipRateLimiter)

		if ipLimiter.limiter.TakeAvailable(1) == 0 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		ipLimiter.lastSeen = time.Now()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanupRateLimiters() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.ipRateLimiters.Range(func(key, value interface{}) bool {
			ip := key.(string)
			ipLimiter := value.(*ipRateLimiter)
			if time.Since(ipLimiter.lastSeen) > rl.config.CleanupInterval {
				rl.ipRateLimiters.Delete(ip)
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
		// Handle error (log it or return a default value)
		// Log the error for debugging.
		log.Printf("Error splitting RemoteAddr: %v", err)
		return "0.0.0.0" // Or a more appropriate default
	}
	return ip
}
