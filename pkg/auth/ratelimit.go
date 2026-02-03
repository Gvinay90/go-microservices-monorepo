package auth

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RateLimiter implements token bucket rate limiting per user
type RateLimiter struct {
	buckets map[string]*bucket
	mu      sync.RWMutex
	rate    int           // requests per window
	window  time.Duration // time window
	log     *slog.Logger
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window duration
func NewRateLimiter(rate int, window time.Duration, log *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
		log:     log,
	}

	// Cleanup old buckets periodically
	go rl.cleanup()

	return rl
}

// Interceptor creates a rate limiting interceptor
func (rl *RateLimiter) Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract user ID from context (set by auth interceptor)
		userID := "anonymous"
		if claims, ok := GetClaims(ctx); ok {
			userID = GetUserID(claims)
		} else {
			// For unauthenticated requests, use IP address
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
					userID = ips[0]
				}
			}
		}

		// Check rate limit
		if !rl.allow(userID) {
			rl.log.Warn("Rate limit exceeded", "user", userID, "method", info.FullMethod)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
		}

		return handler(ctx, req)
	}
}

// allow checks if a request is allowed for the given user
func (rl *RateLimiter) allow(userID string) bool {
	rl.mu.RLock()
	b, exists := rl.buckets[userID]
	rl.mu.RUnlock()

	if !exists {
		// Create new bucket
		b = &bucket{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.mu.Lock()
		rl.buckets[userID] = b
		rl.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens if window has passed
	now := time.Now()
	if now.Sub(b.lastRefill) >= rl.window {
		b.tokens = rl.rate
		b.lastRefill = now
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes old buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for userID, b := range rl.buckets {
			b.mu.Lock()
			if now.Sub(b.lastRefill) > rl.window*2 {
				delete(rl.buckets, userID)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
