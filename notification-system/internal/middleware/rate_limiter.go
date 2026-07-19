package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter stores independent token buckets for every unique visitor IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new rate limiter allowing 'r' requests per second with a burst of 'b'.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	// Note: In a true production environment, I would also spawn a background goroutine 
	// to clean up old IPs from the map to prevent a slow memory leak over time.
}

// AddIP safely creates a brand new Token Bucket for an unseen IP
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter
	return limiter
}

// GetLimiter safely retrieves the Token Bucket for the provided IP, creating one if it doesn't exist
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	return limiter
}

// RateLimit is the actual Gin middleware function that intercepts requests
func RateLimit(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiterForIP := limiter.GetLimiter(ip)

		// Allow() mathematically removes exactly 1 token from the bucket.
		// If the bucket is completely empty, it returns false.
		if !limiterForIP.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please slow down.",
			})
			c.Abort() // 🚨 Instantly stops the request from reaching the API handlers
			return
		}

		c.Next()
	}
}
