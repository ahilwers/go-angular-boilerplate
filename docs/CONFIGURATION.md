# Configuration

Configuration uses [**Viper**](https://github.com/spf13/viper) library and is loaded with the following precedence (highest to lowest):
1. Environment variables (highest priority)
2. Config file: `config/local.yaml` (default) or path specified in `CONFIG_PATH` env var
3. Default values (lowest priority)

Viper automatically handles:
- Multiple config formats (YAML, JSON, TOML)
- Environment variable mapping (e.g., `SERVICE_PORT` → `service.port`)
- Config file watching (hot reload capability)
- Nested configuration structures

## Key Configuration Options

### Service
- `SERVICE_HOST`: HTTP server host (default: localhost)
- `SERVICE_PORT`: HTTP server port (default: 8080)
- `SERVICE_READ_TIMEOUT`: Read timeout in seconds (default: 10)
- `SERVICE_WRITE_TIMEOUT`: Write timeout in seconds (default: 10)

### Database
- `DATABASE_URI`: MongoDB connection string (default: mongodb://localhost:27017)
- `DATABASE_NAME`: Database name (default: boilerplate)
- `DATABASE_USERNAME`: MongoDB username (optional, for authenticated connections)
- `DATABASE_PASSWORD`: MongoDB password (optional, for authenticated connections)

### Authentication
- `AUTH_ENABLED`: Enable/disable authentication (default: false)
- `AUTH_ISSUER`: OpenID issuer URL
- `AUTH_CLIENT_ID`: OAuth client ID
- `AUTH_CLIENT_SECRET`: OAuth client secret
- `AUTH_JWKS_URL`: JWKS endpoint for token validation

### Logging
- `LOG_LEVEL`: Log level (debug, info, warn, error)
- `LOG_FORMAT`: Output format (console, json)
- `LOKI_URL`: Loki push endpoint (optional)
- `LOKI_BEARER_TOKEN`: Loki authentication token (optional)

### API Documentation
- `DOCS_ENABLED`: Enable/disable API documentation endpoints (default: true)
  - Set to `true` in development for easy API testing
  - Set to `false` in production for security (hides API surface)
