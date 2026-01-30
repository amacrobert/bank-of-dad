package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old entries
	entries := rl.attempts[key]
	valid := entries[:0]
	for _, t := range entries {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.attempts[key] = valid
		return false
	}

	rl.attempts[key] = append(valid, now)
	return true
}

// RateLimit returns middleware that limits requests per IP address.
// limit is the max number of requests allowed within the window duration.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = fwd
			}

			if !rl.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too many requests","message":"Please wait before trying again."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
