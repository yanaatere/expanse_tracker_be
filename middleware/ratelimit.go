package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// rateLimitScript atomically increments the request counter and sets the TTL on
// the first increment. Using a Lua script ensures INCR + EXPIRE are atomic.
var rateLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// RateLimit returns a middleware that allows at most limit requests per window
// per client IP for the given routeKey. On Redis failure it fails open so that
// a Redis outage never takes down the API.
func RateLimit(client *redis.Client, routeKey string, limit int, window time.Duration) func(http.Handler) http.Handler {
	windowSecs := strconv.Itoa(int(window.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			key := fmt.Sprintf("ratelimit:%s:%s", routeKey, ip)

			count, err := rateLimitScript.Run(context.Background(), client, []string{key}, windowSecs).Int64()
			if err != nil {
				// Fail open — don't block users on Redis outage.
				next.ServeHTTP(w, r)
				return
			}

			if count > int64(limit) {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the real client IP, trusting X-Forwarded-For when present.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; the first entry is the client.
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
