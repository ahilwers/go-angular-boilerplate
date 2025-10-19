package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		shouldPanic    bool
	}{
		{
			name: "handler that panics with string",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
			shouldPanic:    false,
		},
		{
			name: "handler that panics with error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrAbortHandler)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
			shouldPanic:    false,
		},
		{
			name: "handler that panics with nil",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var ptr *string
				_ = *ptr // nil pointer dereference
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
			shouldPanic:    false,
		},
		{
			name: "handler that does not panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			shouldPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger that writes to a buffer so we can verify logging
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			// Create the middleware
			recovery := RecoveryMiddleware(logger)
			wrappedHandler := recovery(tt.handler)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Execute the handler
			wrappedHandler.ServeHTTP(rec, req)

			// Verify status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// Verify response body contains expected text
			if !strings.Contains(rec.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, rec.Body.String())
			}

			// For panic cases, verify that error was logged
			if tt.expectedStatus == http.StatusInternalServerError && tt.name != "handler that does not panic" {
				logOutput := logBuf.String()
				if !strings.Contains(logOutput, "panic recovered in HTTP handler") {
					t.Error("expected panic to be logged")
				}
				if !strings.Contains(logOutput, "level=ERROR") {
					t.Error("expected error level in log")
				}
			}
		})
	}
}

func TestRecoveryMiddleware_PreservesHeaders(t *testing.T) {
	// Test that recovery middleware doesn't interfere with normal response headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	recovery := RecoveryMiddleware(logger)
	wrappedHandler := recovery(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if rec.Header().Get("X-Custom-Header") != "test-value" {
		t.Error("expected custom header to be preserved")
	}

	if rec.Body.String() != "created" {
		t.Errorf("expected body %q, got %q", "created", rec.Body.String())
	}
}

func TestRecoveryMiddleware_PanicInMiddlewareChain(t *testing.T) {
	// Test that recovery middleware catches panics from other middleware
	panicMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("middleware panic")
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	})

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	// Chain: Recovery -> PanicMiddleware -> FinalHandler
	recovery := RecoveryMiddleware(logger)
	handler := recovery(panicMiddleware(finalHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Internal Server Error") {
		t.Error("expected error message in response")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "panic recovered in HTTP handler") {
		t.Error("expected panic to be logged")
	}
	if !strings.Contains(logOutput, "middleware panic") {
		t.Error("expected panic message in log")
	}
}
