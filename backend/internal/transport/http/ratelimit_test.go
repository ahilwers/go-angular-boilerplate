package http

import (
	"boilerplate/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Middleware(t *testing.T) {
	tests := []struct {
		name               string
		rps                int
		burst              int
		requestCount       int
		requestDelay       time.Duration
		expectedSuccessful int
		expectedBlocked    int
	}{
		{
			name:               "allows requests within limit",
			rps:                10,
			burst:              20,
			requestCount:       5,
			requestDelay:       0,
			expectedSuccessful: 5,
			expectedBlocked:    0,
		},
		{
			name:               "blocks requests exceeding burst",
			rps:                1,
			burst:              2,
			requestCount:       5,
			requestDelay:       0,
			expectedSuccessful: 2, // burst allows 2 requests
			expectedBlocked:    3,
		},
		{
			name:               "allows requests with delay",
			rps:                10,
			burst:              1,
			requestCount:       3,
			requestDelay:       110 * time.Millisecond, // slightly more than 1/10 second
			expectedSuccessful: 3,
			expectedBlocked:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create rate limiter with test config
			cfg := config.RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: tt.rps,
				Burst:             tt.burst,
			}
			rateLimiter := NewRateLimiter(cfg)

			// Create a simple handler that always returns 200
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with rate limiting middleware
			limitedHandler := rateLimiter.Middleware()(handler)

			successful := 0
			blocked := 0

			// Make requests
			for i := 0; i < tt.requestCount; i++ {
				if i > 0 && tt.requestDelay > 0 {
					time.Sleep(tt.requestDelay)
				}

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "127.0.0.1:1234" // Same IP for all requests
				rec := httptest.NewRecorder()

				limitedHandler.ServeHTTP(rec, req)

				if rec.Code == http.StatusOK {
					successful++
				} else if rec.Code == http.StatusTooManyRequests {
					blocked++
				}
			}

			if successful != tt.expectedSuccessful {
				t.Errorf("expected %d successful requests, got %d", tt.expectedSuccessful, successful)
			}

			if blocked != tt.expectedBlocked {
				t.Errorf("expected %d blocked requests, got %d", tt.expectedBlocked, blocked)
			}
		})
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	// Create rate limiter with strict limits
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerSecond: 1,
		Burst:             1,
	}
	rateLimiter := NewRateLimiter(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limitedHandler := rateLimiter.Middleware()(handler)

	// Make request from first IP
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rec1 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("expected first request from IP1 to succeed, got status %d", rec1.Code)
	}

	// Make request from second IP (should succeed even though first IP exhausted limit)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	rec2 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("expected first request from IP2 to succeed, got status %d", rec2.Code)
	}

	// Make another request from first IP (should fail due to rate limit)
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	rec3 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("expected second request from IP1 to be blocked, got status %d", rec3.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "uses RemoteAddr when no headers",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1:1234",
		},
		{
			name:          "prefers X-Forwarded-For",
			remoteAddr:    "192.168.1.1:1234",
			xForwardedFor: "10.0.0.1",
			xRealIP:       "10.0.0.2",
			expectedIP:    "10.0.0.1",
		},
		{
			name:       "uses X-Real-IP when X-Forwarded-For is absent",
			remoteAddr: "192.168.1.1:1234",
			xRealIP:    "10.0.0.2",
			expectedIP: "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}
