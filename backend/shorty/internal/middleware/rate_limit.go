package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"shorty/internal/utils"
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*tokenBucket
	capacity   float64
	refillRate float64
	ttl        time.Duration
}

func NewRateLimiter(maxPerMinute int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*tokenBucket),
		capacity:   float64(maxPerMinute),
		refillRate: float64(maxPerMinute) / 60.0,
		ttl:        ttl,
	}
}

func (rl *RateLimiter) Allow(ip string) (allowed bool, retryAfterSec int) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &tokenBucket{
			tokens:     rl.capacity - 1,
			lastRefill: now,
		}
		return true, 0
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens = minFloat(rl.capacity, b.tokens+elapsed*rl.refillRate)
		b.lastRefill = now
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true, 0
	}

	need := 1 - b.tokens
	sec := int((need/rl.refillRate)+0.999)
	if sec < 1 {
		sec = 1
	}
	return false, sec
}

func (rl *RateLimiter) CleanupLoop(stop <-chan struct{}) {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			rl.cleanup()
		case <-stop:
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, b := range rl.buckets {
		if now.Sub(b.lastRefill) > rl.ttl {
			delete(rl.buckets, ip)
		}
	}
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.GetClientIP(r)

			allowed, retryAfter := rl.Allow(ip)
			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				http.Error(w, "rate limit exceeded (max 10 requests/minute)", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
