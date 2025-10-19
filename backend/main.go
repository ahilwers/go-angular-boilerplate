package main

import (
	"boilerplate/internal/auth"
	"boilerplate/internal/config"
	"boilerplate/internal/logger"
	"boilerplate/internal/service"
	"boilerplate/internal/storage"
	httpTransport "boilerplate/internal/transport/http"
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// @title           Boilerplate API
// @version         1.0
// @description     Production-ready full-stack todo application API with Go backend and MongoDB persistence
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.oauth2.implicit BearerAuth
// @authorizationUrl http://localhost:8081/realms/boilerplate/protocol/openid-connect/auth
// @tokenUrl http://localhost:8081/realms/boilerplate/protocol/openid-connect/token
// @scope.openid OpenID Connect scope
// @scope.profile Profile scope
// @scope.email Email scope

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("Warning: failed to load config file: %v. Using defaults and environment variables.", err)
		// Try loading with empty path to use defaults
		cfg, err = config.Load("")
		if err != nil {
			log.Fatalf("Failed to initialize configuration: %v", err)
		}
	}

	appLogger := logger.New(cfg.Logging)
	slog.SetDefault(appLogger)

	appLogger.Info("starting boilerplate server",
		"service_host", cfg.Service.Host,
		"service_port", cfg.Service.Port,
		"auth_enabled", cfg.Auth.Enabled,
	)

	appLogger.Info("connecting to MongoDB", "uri", cfg.Database.URI)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.Timeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.Database.URI)

	// Add authentication if credentials are provided
	if cfg.Database.Username != "" && cfg.Database.Password != "" {
		credential := options.Credential{
			Username: cfg.Database.Username,
			Password: cfg.Database.Password,
		}
		clientOptions.SetAuth(credential)
		appLogger.Info("MongoDB authentication enabled", "username", cfg.Database.Username)
	}

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		appLogger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		appLogger.Error("failed to ping MongoDB", "error", err)
		os.Exit(1)
	}
	appLogger.Info("connected to MongoDB")

	repo := storage.NewRepository(mongoClient, cfg.Database.Database)
	svc := service.NewService(&repo)

	authMiddleware := auth.NewMiddleware(cfg.Auth, appLogger)

	httpServer := httpTransport.NewServer(cfg.Service, cfg.CORS, cfg.Auth, cfg.Docs, cfg.RateLimit, svc, authMiddleware, appLogger)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.Start()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		appLogger.Error("server error", "error", err)
	case sig := <-shutdown:
		appLogger.Info("received shutdown signal", "signal", sig)

		// Graceful shutdown with 30 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			appLogger.Error("failed to gracefully shutdown server", "error", err)
			if err := mongoClient.Disconnect(context.Background()); err != nil {
				appLogger.Error("failed to disconnect from MongoDB", "error", err)
			}
			os.Exit(1)
		}

		if err := mongoClient.Disconnect(ctx); err != nil {
			appLogger.Error("failed to disconnect from MongoDB", "error", err)
		}

		appLogger.Info("server shutdown complete")
	}
}
