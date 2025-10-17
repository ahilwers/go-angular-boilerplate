# go-angular-boilerplate

Production-ready monorepo boilerplate with a Go backend and Angular frontend. Features clean hexagonal architecture, Keycloak authentication, structured logging, and testing.

## Features

### Backend
- **Clean Architecture**: Hexagonal/ports & adapters pattern with clear separation of concerns
- **RESTful API**: Full CRUD operations for projects and tasks under `/api/v1`
- **MongoDB Persistence**: Official MongoDB driver with repository pattern
- **Authentication**: Keycloak/OpenID Connect with JWT validation and JWKS support
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

Configuration uses [**Viper**](https://github.com/spf13/viper) library and is loaded with the following precedence (highest to lowest):
1. Environment variables (highest priority)
2. Config file: `config/local.yaml` (default) or path specified in `CONFIG_PATH` env var
3. Default values (lowest priority)

Viper automatically handles:
- Multiple config formats (YAML, JSON, TOML)
- Environment variable mapping (e.g., `SERVICE_PORT` → `service.port`)
- Config file watching (hot reload capability)
- Nested configuration structures

### Key Configuration Options

#### Service
- `SERVICE_HOST`: HTTP server host (default: localhost)
- `SERVICE_PORT`: HTTP server port (default: 8080)
- `SERVICE_READ_TIMEOUT`: Read timeout in seconds (default: 10)
- `SERVICE_WRITE_TIMEOUT`: Write timeout in seconds (default: 10)

#### Database
- `DATABASE_URI`: MongoDB connection string (default: mongodb://localhost:27017)
- `DATABASE_NAME`: Database name (default: boilerplate)
- `DATABASE_USERNAME`: MongoDB username (optional, for authenticated connections)
- `DATABASE_PASSWORD`: MongoDB password (optional, for authenticated connections)

#### Authentication
- `AUTH_ENABLED`: Enable/disable authentication (default: false)
- `AUTH_ISSUER`: OpenID issuer URL
- `AUTH_CLIENT_ID`: OAuth client ID
- `AUTH_CLIENT_SECRET`: OAuth client secret
- `AUTH_JWKS_URL`: JWKS endpoint for token validation

#### Logging
- `LOG_LEVEL`: Log level (debug, info, warn, error)
- `LOG_FORMAT`: Output format (console, json)
- `LOKI_URL`: Loki push endpoint (optional)
- `LOKI_BEARER_TOKEN`: Loki authentication token (optional)

## API Documentation

### Projects

#### List all projects
```bash
GET /api/v1/projects
```

#### Get a project
```bash
GET /api/v1/projects/{id}
```

#### Create a project
```bash
POST /api/v1/projects
Content-Type: application/json

{
  "name": "My Project",
  "description": "Project description"
}
```

#### Update a project
```bash
PUT /api/v1/projects/{id}
Content-Type: application/json

{
  "name": "Updated Project",
  "description": "Updated description"
}
```

#### Delete a project
```bash
DELETE /api/v1/projects/{id}
```

### Tasks

#### List tasks for a project
```bash
GET /api/v1/projects/{id}/tasks
```

#### Get a task
```bash
GET /api/v1/tasks/{id}
```

#### Create a task
```bash
POST /api/v1/projects/{projectId}/tasks
Content-Type: application/json

{
  "title": "My Task",
  "status": "TODO",
  "description": "Task description",
  "due_date": "2025-12-31T00:00:00Z"
}
```

Task status can be: `TODO`, `IN_PROGRESS`, or `DONE`

#### Update a task
```bash
PUT /api/v1/tasks/{id}
Content-Type: application/json

{
  "title": "Updated Task",
  "status": "IN_PROGRESS",
  "description": "Updated description"
}
```

#### Delete a task
```bash
DELETE /api/v1/tasks/{id}
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
