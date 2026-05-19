package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Store request timestamps per user
type UserLimiter struct {
	requests []time.Time
	mu       sync.Mutex
}

// Global map of user limiters
var (
	limiters = make(map[uint]*UserLimiter)
	mu       sync.Mutex
)

const (
	MAX_REQUESTS = 20             // max requests
	WINDOW       = time.Hour      // per hour
	// MAX_REQUESTS = 3  // change to 3 for testing
	// WINDOW = time.Minute // change to 1 minute
)

func getUserLimiter(userID uint) *UserLimiter {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := limiters[userID]; !exists {
		limiters[userID] = &UserLimiter{}
	}
	return limiters[userID]
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		// Get or create limiter for this user
		limiter := getUserLimiter(userID)
		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		now := time.Now()
		windowStart := now.Add(-WINDOW)

		// Remove requests outside the window
		var validRequests []time.Time
		for _, t := range limiter.requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		limiter.requests = validRequests

		// Check if limit exceeded
		if len(limiter.requests) >= MAX_REQUESTS {
			// Calculate when oldest request expires
			oldestRequest := limiter.requests[0]
			resetTime := oldestRequest.Add(WINDOW)
			waitSeconds := int(time.Until(resetTime).Seconds())

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"limit":       MAX_REQUESTS,
				"window":      "1 hour",
				"retry_after": waitSeconds,
				"message":     "Too many requests. Please wait before sending more messages.",
			})
			c.Abort()
			return
		}

		// Add current request
		limiter.requests = append(limiter.requests, now)

		// Add helpful headers
		c.Header("X-RateLimit-Limit", "20")
		c.Header("X-RateLimit-Remaining", string(rune(MAX_REQUESTS-len(limiter.requests))))

		c.Next()
	}
}