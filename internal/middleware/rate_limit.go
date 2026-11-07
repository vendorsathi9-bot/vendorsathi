package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"vendorsaathi/internal/platform/httpx"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
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
			key = userID.(string)
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

		if v.count >= rl.rate {
			rl.mu.Unlock()
			retryAfter := int(rl.window.Seconds()) - int(now.Sub(v.lastSeen).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", string(rune(retryAfter)))
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
