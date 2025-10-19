# go-angular-boilerplate

Production-ready monorepo boilerplate with a Go backend and Angular frontend. Features clean hexagonal architecture, Keycloak authentication, structured logging, and testing.

## Features

### Backend
- **Clean Architecture**: Hexagonal/ports & adapters pattern with clear separation of concerns
- **RESTful API**: Full CRUD operations for projects and tasks under `/api/v1`
- **MongoDB Persistence**: Official MongoDB driver with repository pattern
- **Authentication**: Keycloak/OpenID Connect with JWT validation and JWKS support
- **Rate Limiting**: Per-IP rate limiting using token bucket algorithm
- **Configuration**: YAML/JSON config files with environment variable overrides
- **Structured Logging**: `slog` with console and Loki handlers
- **Comprehensive Tests**: Unit tests with mocks and integration tests with Testcontainers

### Frontend
- **Angular 18+**: Modern Angular with standalone components (planned)
- **PrimeNG UI**: Professional UI component library (planned)
- **System-aware Theming**: Light/dark mode following system preferences (planned)

## Tech Stack

### Backend
- **Go 1.24**: Modern Go
- **MongoDB**: Official Go driver with connection pooling
- **Keycloak**: OpenID Connect authentication (swappable)
- **Structured Logging**: `slog` with Loki integration
- **Testing**: Go testing, Testcontainers, httptest

### Frontend (Planned)
- **Angular**: Signals-ready architecture
- **PrimeNG**: Enterprise UI components
- **RxJS**: Reactive programming

## Project Structure

```
.
├── backend/
│   ├── internal/
│   │   ├── entities/        # Domain models (Project, Task)
│   │   ├── service/         # Business logic layer
│   │   │   └── domain/      # Service implementations
│   │   ├── storage/         # Repository interfaces
│   │   │   └── mongodb/     # MongoDB implementations
│   │   ├── transport/       # HTTP handlers and routing
│   │   │   └── http/        # REST API handlers
│   │   ├── auth/            # Authentication middleware
│   │   ├── config/          # Configuration management
│   │   └── logger/          # Logging utilities
│   ├── main.go              # Application entry point
│   └── go.mod
├── config/
│   └── local.yaml           # Local development config
├── keycloak/                # Keycloak setup docs
├── docker-compose.yaml      # Docker services
├── Makefile                 # Development commands
└── README.md
```

## Quick Start

### Prerequisites
- Go 1.24 or later
- Docker and Docker Compose
- Make (optional, but recommended)

### 1. Clone the repository
```bash
git clone <repository-url>
cd go-angular-boilerplate
```

### 2. Start infrastructure services
```bash
make docker-up
# OR
docker-compose up -d mongodb keycloak
```

This starts:
- MongoDB on port 27017
- Keycloak on port 8081

### 3. Configure environment (optional)
```bash
cp .env.example .env
# Edit .env as needed
```

### 4. Run the backend server
```bash
make run
# OR
cd backend && CONFIG_PATH=../config/local.yaml go run main.go
```

The server will start on http://localhost:8080

### 5. Verify it's working
```bash
# Health check
curl http://localhost:8080/health

# List projects (auth disabled by default)
curl http://localhost:8080/api/v1/projects
```

## Development

### Running tests
```bash
# All tests
make test

# Unit tests only (fast)
make test-unit

# Integration tests (requires Docker)
make test-integration
```

### Code formatting and linting
```bash
make fmt    # Format code
make lint   # Run linters
```

### Building
```bash
make build  # Creates backend/bin/server
```

## Configuration

**📖 See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for complete configuration options!**

Quick overview:
- Uses [Viper](https://github.com/spf13/viper) for flexible configuration management
- Environment variables override config file values
- Default config: `config/local.yaml` or set `CONFIG_PATH` env var
- Supports multiple formats: YAML, JSON, TOML

## API Documentation

**📖 See [backend/docs/README.md](backend/docs/README.md) for complete API documentation!**

The API documentation is auto-generated using OpenAPI/Swagger and available through:
- **Scalar UI** (recommended): http://localhost:8080/docs/scalar - Modern, interactive API docs with OAuth2 integration
- **Swagger UI**: http://localhost:8080/swagger/index.html - Traditional Swagger interface
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json - Raw JSON specification

### Quick API Examples

```bash
# List all projects
GET /api/v1/projects

# Create a project
POST /api/v1/projects
Content-Type: application/json
{"name": "My Project", "description": "Project description"}

# Create a task
POST /api/v1/projects/{projectId}/tasks
Content-Type: application/json
{"title": "My Task", "status": "TODO", "description": "Task description"}
```

Task status values: `TODO`, `IN_PROGRESS`, `DONE`

## Rate Limiting

The API includes per-IP rate limiting to protect against abuse and ensure fair resource usage.

### Configuration

By default, rate limiting is **enabled** with the following settings:
- **Requests per second**: 10
- **Burst size**: 20

Update `config/local.yaml` to customize:
```yaml
rate_limit:
  enabled: true
  requests_per_second: 10  # Sustained rate limit
  burst: 20                # Allows temporary spikes
```

Or use environment variables:
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST=20
```

### How it works

- Uses the **token bucket algorithm** via `golang.org/x/time/rate`
- Rate limiting is applied **per IP address**
- Supports `X-Forwarded-For` and `X-Real-IP` headers for proxies
- Returns `429 Too Many Requests` when limit is exceeded
- Applied to all `/api/*` routes (health checks are exempt)

### Disabling rate limiting

For development or testing, you can disable rate limiting:
```yaml
rate_limit:
  enabled: false
```

## Authentication

By default, authentication is **disabled** for easier development. To enable Keycloak authentication:

### 1. Configure Keycloak

**📖 See [keycloak/README.md](keycloak/README.md) for complete step-by-step instructions!**

Quick steps:
1. Start Keycloak: `docker-compose up -d keycloak`
2. Access Keycloak at http://localhost:8081 (admin/admin)
3. Create realm: `boilerplate`
4. Create client: `boilerplate-client` (Client authentication: ON)
5. Copy client secret from Credentials tab
6. Create test users

### 2. Enable authentication

Update `config/local.yaml`:
```yaml
auth:
  enabled: true
  issuer: "http://localhost:8081/realms/boilerplate"
  client_id: "boilerplate-client"
  client_secret: "YOUR_CLIENT_SECRET"
  jwks_url: "http://localhost:8081/realms/boilerplate/protocol/openid-connect/certs"
```

### 3. Get an access token

```bash
curl -X POST http://localhost:8081/realms/boilerplate/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=boilerplate-client" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "grant_type=password" \
  -d "username=testuser" \
  -d "password=testpass"
```

### 4. Use token in API requests

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/v1/projects
```

## Docker Services

### Start all services
```bash
make docker-up
```

### Start with observability (Loki + Grafana)
```bash
make docker-up-full
```

Services:
- MongoDB: http://localhost:27017
- Keycloak: http://localhost:8081
- Loki: http://localhost:3100 (with observability profile)
- Grafana: http://localhost:3000 (with observability profile)

### View logs
```bash
make docker-logs
```

### Stop services
```bash
make docker-down
```

## Testing Strategy

### Unit Tests
- Service layer with mocked repositories
- HTTP handlers with mocked services
- Run with: `make test-unit`

### Integration Tests
- Repository layer with Testcontainers MongoDB
- Guarded with `testing.Short()` to skip in CI
- Run with: `make test-integration`

### API Contract Tests
- HTTP handlers with httptest
- Validates request/response formats
- Run with: `make test`

## Architecture

The backend follows hexagonal (ports & adapters) architecture:

1. **Entities** (`internal/entities/`): Domain models
2. **Services** (`internal/service/`): Business logic
3. **Storage** (`internal/storage/`): Repository interfaces and implementations
4. **Transport** (`internal/transport/`): HTTP handlers and routing
5. **Auth** (`internal/auth/`): Authentication middleware
6. **Config** (`internal/config/`): Configuration management
7. **Logger** (`internal/logger/`): Logging utilities

### Data Flow
```
HTTP Request → Auth Middleware → Handler → Service → Repository → MongoDB
```

### Dependency Injection
- Repositories are injected into services
- Services are injected into handlers
- Middleware wraps handlers

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linters: `make lint`
6. Submit a pull request

## License

See LICENSE file for details.
