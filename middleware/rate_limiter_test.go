package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRoute creates a gin engine with a single /test endpoint protected by the given limiter.
func setupTestRoute(limiter gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", limiter, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 3,
		Window:   time.Minute,
	})
	r := setupTestRoute(limiter)

	// Send 3 requests — all should succeed
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_BlocksWhenLimitExceeded(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 2,
		Window:   time.Minute,
	})
	r := setupTestRoute(limiter)

	// First 2 requests succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should be allowed", i+1)
	}

	// 3rd request should be blocked with 429
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "excess request should be blocked")

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "rate limit exceeded")
}

func TestRateLimiter_DifferentIPsHaveIndependentLimits(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 1,
		Window:   time.Minute,
	})
	r := setupTestRoute(limiter)

	// First IP uses its request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "first IP first request should be allowed")

	// Same IP is blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code, "same IP should be blocked")

	// Different IP is allowed (independent counter)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "10.0.0.1:54321"
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code, "different IP should be allowed")
}

func TestRateLimiter_WindowResetsAfterTime(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 1,
		Window:   50 * time.Millisecond,
	})
	r := setupTestRoute(limiter)

	// First request succeeds
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request (still within window) is blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Wait for the window to expire
	time.Sleep(60 * time.Millisecond)

	// Request should succeed again (window has passed)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code, "request should be allowed after window passes")
}

func TestRateLimiter_ZeroConfigDefaults(t *testing.T) {
	// Both zero values should default to Requests=1, Window=1 minute
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 0,
		Window:   0,
	})
	r := setupTestRoute(limiter)

	// First request succeeds
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request is blocked (default limit = 1)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_HandlerCalledOnSuccess(t *testing.T) {
	// Verify the downstream handler is actually invoked when the request is allowed
	called := false
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter(RateLimiterConfig{
		Requests: 1,
		Window:   time.Minute,
	})
	r := gin.New()
	r.GET("/test", limiter, func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.True(t, called, "downstream handler should be called when under limit")
	assert.Equal(t, http.StatusOK, w.Code)
}


