package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig defines the parameters for a rate limiter instance.
type RateLimiterConfig struct {
	Requests int           // maximum number of requests allowed within the window
	Window   time.Duration // sliding window duration
}

// rateLimiter implements a per-IP sliding window rate limiter.
type rateLimiter struct {
	mu       sync.RWMutex
	visitors map[string][]time.Time
	config   RateLimiterConfig
}

// NewRateLimiter creates a Gin middleware that rate-limits requests per client IP
// using a sliding window. Each call creates an independent limiter instance.
func NewRateLimiter(config RateLimiterConfig) gin.HandlerFunc {
	if config.Requests <= 0 {
		config.Requests = 1
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}

	rl := &rateLimiter{
		visitors: make(map[string][]time.Time),
		config:   config,
	}

	// Periodically purge stale entries to avoid memory leaks
	go rl.cleanup()

	return rl.handler
}

// handler processes each incoming request.
func (rl *rateLimiter) handler(c *gin.Context) {
	ip := c.ClientIP()

	rl.mu.Lock()
	now := time.Now()
	cutoff := now.Add(-rl.config.Window)

	// Retain only timestamps that fall within the sliding window
	timestamps := rl.visitors[ip]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	// Check if the client has exceeded the limit
	if len(valid) >= rl.config.Requests {
		rl.mu.Unlock()
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded, please try again later",
		})
		return
	}

	// Record this request
	valid = append(valid, now)
	rl.visitors[ip] = valid
	rl.mu.Unlock()

	c.Next()
}

// cleanup runs periodically to remove stale entries from the visitors map.
func (rl *rateLimiter) cleanup() {
	interval := rl.config.Window / 2
	if interval < time.Second {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.config.Window)
		for ip, timestamps := range rl.visitors {
			var valid []time.Time
			for _, t := range timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.visitors, ip)
			} else {
				rl.visitors[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}
