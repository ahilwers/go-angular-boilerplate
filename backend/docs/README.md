# API Documentation

This directory contains auto-generated OpenAPI/Swagger documentation for the Boilerplate API.

## Overview

The API documentation is generated using [swaggo/swag](https://github.com/swaggo/swag) and served via:
- **Scalar UI** (recommended): Modern, interactive API documentation with OAuth2 integration
- **Swagger UI**: Traditional Swagger interface

## Accessing the Documentation

Once the backend server is running (on `http://localhost:8080` by default), you can access:

### Scalar UI (Recommended)
- **URL**: http://localhost:8080/docs/scalar
- **Features**:
  - Beautiful, modern interface
  - Integrated OAuth2 authentication with Keycloak
  - Try out API endpoints directly from the UI
  - Automatic token management
  - Dark/light theme support

### Swagger UI
- **URL**: http://localhost:8080/swagger/index.html
- **Features**: Traditional Swagger interface

### Raw OpenAPI Specification
- **JSON**: http://localhost:8080/swagger/doc.json
- **YAML**: http://localhost:8080/swagger/swagger.yaml

## OAuth2 Authentication in Scalar

The Scalar UI is pre-configured with Keycloak OAuth2 authentication:

1. Click the "Authenticate" button in Scalar UI
2. You'll be redirected to Keycloak login page
3. Login with your Keycloak credentials (default: username/password from Keycloak setup)
4. After successful authentication, you'll be redirected back to Scalar
5. The access token is automatically included in all API requests

### OAuth2 Configuration
- **Authorization URL**: http://localhost:8081/realms/boilerplate/protocol/openid-connect/auth
- **Token URL**: http://localhost:8081/realms/boilerplate/protocol/openid-connect/token
- **Client ID**: boilerplate-client
- **Scopes**: openid, profile, email

## Regenerating Documentation

After modifying API handlers or adding new endpoints:

```bash
# From the backend directory
swag init --output docs --parseDependency --parseInternal

# Or from the project root
make docs
```

This will regenerate:
- `docs/docs.go` - Embedded Go documentation
- `docs/swagger.json` - OpenAPI specification (JSON)
- `docs/swagger.yaml` - OpenAPI specification (YAML)

## Adding Documentation to Endpoints

Use Swagger annotations in your Go code:

```go
// List godoc
// @Summary      List projects
// @Description  Get all projects with optional pagination
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        page   query  int  false  "Page number (1-based)"
// @Param        limit  query  int  false  "Items per page"
// @Success      200  {array}   entities.Project
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/projects [get]
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
    // handler implementation
}
```

## Files in This Directory

- `docs.go` - Generated Go code with embedded OpenAPI spec
- `swagger.json` - OpenAPI 2.0 specification in JSON format
- `swagger.yaml` - OpenAPI 2.0 specification in YAML format
- `README.md` - This file

## Customization

### Scalar UI Theme
Edit `internal/transport/http/docs_handler.go` to customize Scalar appearance:

```javascript
"theme": "purple",        // or "blue", "green", "default"
"layout": "modern",       // or "classic"
"defaultOpenAllTags": false
```

### API Metadata
Edit the main annotations in `main.go`:

```go
// @title           Boilerplate API
// @version         1.0
// @description     Production-ready full-stack todo application API
// @host            localhost:8080
// @BasePath        /
```

## Integration with CI/CD

To ensure documentation stays up-to-date:

1. Add a pre-commit hook to regenerate docs:
   ```bash
   #!/bin/bash
   make docs
   git add backend/docs/
   ```

2. Add a CI check to verify docs are current:
   ```bash
   make docs
   git diff --exit-code backend/docs/
   ```

## Troubleshooting

### Documentation not updating
- Make sure you run `make docs` or `swag init` after code changes
- Restart the backend server to reload the embedded docs

### OAuth2 not working
- Verify Keycloak is running: `docker ps | grep keycloak`
- Check Keycloak realm configuration matches the client ID
- Ensure the redirect URIs are configured in Keycloak client settings

### 404 on documentation endpoints
- Verify the server is running
- Check that `docs` package is imported in `main.go` with `_ "boilerplate/docs"`
- Ensure handlers are registered in `internal/transport/http/server.go`
