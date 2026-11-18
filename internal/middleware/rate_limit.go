package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/platform/httpx"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	burst    int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	burst := rate / 4
	if burst < 5 {
		burst = 5
	}
	return NewRateLimiterWithBurst(rate, burst, window)
}

func NewRateLimiterWithBurst(rate, burst int, window time.Duration) *RateLimiter {
	if burst < 0 {
		burst = 0
	}
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok && id != "" {
				key = "user:" + id
			}
		}

		rl.mu.Lock()
		v, exists := rl.visitors[key]
		now := time.Now()
		if !exists || now.Sub(v.lastSeen) > rl.window {
			rl.visitors[key] = &visitor{lastSeen: now, count: 1}
			rl.mu.Unlock()
			c.Next()
			return
		}

		maxAllowed := rl.rate + rl.burst
		if v.count >= maxAllowed {
			rl.mu.Unlock()
			retryAfter := int(rl.window.Seconds()) - int(now.Sub(v.lastSeen).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			httpx.ErrorTooManyRequests(c, "rate limit exceeded")
			c.Abort()
			return
		}

		v.count++
		v.lastSeen = now
		rl.mu.Unlock()
		c.Next()
	}
}
