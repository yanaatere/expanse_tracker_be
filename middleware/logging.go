package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/yanaatere/expense_tracking/logger"
)

// ResponseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// LoggingMiddleware logs all incoming requests and outgoing responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Log incoming request
		logger.Infof(
			"[REQUEST] %s %s | Remote: %s | Query: %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			r.URL.RawQuery,
		)

		// Log request headers (excluding sensitive headers)
		if contentType := r.Header.Get("Content-Type"); contentType != "" {
			logger.Debugf("  Content-Type: %s", contentType)
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			// Don't log the actual token, just that it's present
			logger.Debugf("  Authorization: Bearer [token present]")
		}

		// Wrap response writer to capture status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(startTime)

		// Log response
		statusColor := getStatusColor(rw.statusCode)
		logger.Infof(
			"%s[RESPONSE] %s %s | Status: %d | Duration: %dms%s",
			statusColor,
			r.Method,
			r.RequestURI,
			rw.statusCode,
			duration.Milliseconds(),
			"\033[0m", // Reset color
		)
	})
}

// getStatusColor returns ANSI color code based on status code
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "\033[32m" // Green for 2xx
	case statusCode >= 300 && statusCode < 400:
		return "\033[36m" // Cyan for 3xx
	case statusCode >= 400 && statusCode < 500:
		return "\033[33m" // Yellow for 4xx
	case statusCode >= 500:
		return "\033[31m" // Red for 5xx
	default:
		return ""
	}
}
