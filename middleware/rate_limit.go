package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

// RateLimiter limits requests per IP
func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	// Cleanup routine to remove old clients
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(r, b)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}

// IPBasedRateLimiter provides different limits for different route groups
func IPBasedRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() } // Disabled for development debugging
}

func AuthRateLimiter() gin.HandlerFunc {
	return RateLimiter(rate.Every(3*time.Second), 3) // 1 request every 3 seconds, 3 burst
}
