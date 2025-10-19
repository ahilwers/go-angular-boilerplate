package http

import (
	"boilerplate/internal/auth"
	"boilerplate/internal/config"
	"boilerplate/internal/service"
	"context"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	server         *http.Server
	logger         *slog.Logger
	authMiddleware *auth.Middleware
}

func NewServer(cfg config.ServiceConfig, corsCfg config.CORSConfig, authCfg config.AuthConfig, docsCfg config.DocsConfig, rateLimitCfg config.RateLimitConfig, svc *service.Service, authMw *auth.Middleware, logger *slog.Logger) *Server {
	s := &Server{
		logger:         logger,
		authMiddleware: authMw,
	}

	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)

	// API Documentation endpoints (no auth required for docs)
	// Only register if documentation is enabled in config
	if docsCfg.Enabled {
		logger.Info("API documentation endpoints enabled")
		docsHandler := NewDocsHandler(authCfg)
		mux.HandleFunc("GET /docs", docsHandler.Redirect)
		mux.HandleFunc("GET /docs/scalar", docsHandler.ServeScalar)
		mux.Handle("GET /swagger/", docsHandler.ServeSwaggerUI())
	} else {
		logger.Info("API documentation endpoints disabled")
	}

	// API v1 routes - all protected by auth middleware
	apiMux := http.NewServeMux()

	// Project handlers
	projectHandler := NewProjectHandler(svc.Project, logger)
	apiMux.HandleFunc("GET /api/v1/projects", projectHandler.List)
	apiMux.HandleFunc("POST /api/v1/projects", projectHandler.Create)
	apiMux.HandleFunc("GET /api/v1/projects/{id}", projectHandler.Get)
	apiMux.HandleFunc("PUT /api/v1/projects/{id}", projectHandler.Update)
	apiMux.HandleFunc("DELETE /api/v1/projects/{id}", projectHandler.Delete)

	// Task handlers
	taskHandler := NewTaskHandler(svc.Task, logger)
	apiMux.HandleFunc("GET /api/v1/projects/{id}/tasks", taskHandler.ListByProject)
	apiMux.HandleFunc("POST /api/v1/projects/{id}/tasks", taskHandler.CreateForProject)
	apiMux.HandleFunc("GET /api/v1/tasks/{id}", taskHandler.Get)
	apiMux.HandleFunc("PUT /api/v1/tasks/{id}", taskHandler.Update)
	apiMux.HandleFunc("DELETE /api/v1/tasks/{id}", taskHandler.Delete)

	// Apply middleware chain to API routes: Recovery -> RateLimit -> CORS -> Logging -> Auth
	corsMiddleware := CORSMiddleware(corsCfg)
	recoveryMiddleware := RecoveryMiddleware(logger)
	var apiHandler http.Handler = authMw.Authenticate(apiMux)
	apiHandler = s.loggingMiddleware(apiHandler)
	apiHandler = corsMiddleware(apiHandler)

	// Apply rate limiting if enabled
	if rateLimitCfg.Enabled {
		rateLimiter := NewRateLimiter(rateLimitCfg)
		apiHandler = rateLimiter.Middleware()(apiHandler)
		logger.Info("rate limiting enabled",
			"requests_per_second", rateLimitCfg.RequestsPerSecond,
			"burst", rateLimitCfg.Burst,
		)
	} else {
		logger.Info("rate limiting disabled")
	}

	// Recovery middleware should be outermost to catch all panics
	apiHandler = recoveryMiddleware(apiHandler)

	mux.Handle("/api/", apiHandler)

	s.server = &http.Server{
		Addr:         cfg.Address(),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// TODO: Add checks for database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		s.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", time.Since(start),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
