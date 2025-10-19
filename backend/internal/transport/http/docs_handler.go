package http

import (
	"boilerplate/docs"
	"boilerplate/internal/config"
	"encoding/json"
	"net/http"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/swaggo/swag"
)

// DocsHandler handles API documentation endpoints
type DocsHandler struct {
	authConfig config.AuthConfig
}

// NewDocsHandler creates a new docs handler
func NewDocsHandler(authConfig config.AuthConfig) *DocsHandler {
	return &DocsHandler{
		authConfig: authConfig,
	}
}

// ServeScalar serves the Scalar API documentation UI with Keycloak OAuth2 integration
func (h *DocsHandler) ServeScalar(w http.ResponseWriter, r *http.Request) {
	// Get the swagger spec from swag
	spec := swag.GetSwagger(docs.SwaggerInfo.InstanceName())
	if spec == nil {
		http.Error(w, "Swagger spec not found", http.StatusInternalServerError)
		return
	}

	// Read the swagger spec
	specJSON := spec.ReadDoc()

	// Parse the spec to modify OAuth2 URLs dynamically
	var specMap map[string]interface{}
	if err := json.Unmarshal([]byte(specJSON), &specMap); err != nil {
		http.Error(w, "Failed to parse swagger spec", http.StatusInternalServerError)
		return
	}

	// Update OAuth2 URLs from config if auth is enabled
	if h.authConfig.Enabled {
		if secDefs, ok := specMap["securityDefinitions"].(map[string]interface{}); ok {
			if bearerAuth, ok := secDefs["BearerAuth"].(map[string]interface{}); ok {
				// Update authorization URL from config
				authURL := h.authConfig.Issuer + "/protocol/openid-connect/auth"
				tokenURL := h.authConfig.Issuer + "/protocol/openid-connect/token"

				bearerAuth["authorizationUrl"] = authURL
				if _, hasTokenUrl := bearerAuth["tokenUrl"]; hasTokenUrl {
					bearerAuth["tokenUrl"] = tokenURL
				}
			}
		}
	}

	// Marshal the modified spec back to JSON
	modifiedSpecBytes, err := json.Marshal(specMap)
	if err != nil {
		http.Error(w, "Failed to marshal modified spec", http.StatusInternalServerError)
		return
	}
	specJSON = string(modifiedSpecBytes)

	// Build configuration from auth config
	scalarConfig := map[string]interface{}{
		"theme":              "purple",
		"defaultOpenAllTags": true,
	}

	// Only add authentication config if auth is enabled
	if h.authConfig.Enabled {
		scalarConfig["authentication"] = map[string]interface{}{
			"preferredSecurityScheme": "BearerAuth",
			"oAuth2": map[string]interface{}{
				"clientId": h.authConfig.ClientID,
				"scopes":   []string{"openid", "profile", "email"},
			},
		}
	}

	// Marshal configuration to JSON
	configJSON, err := json.Marshal(scalarConfig)
	if err != nil {
		http.Error(w, "Failed to create configuration", http.StatusInternalServerError)
		return
	}

	html := `<!doctype html>
<html>
  <head>
    <title>Boilerplate API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-configuration='` + string(configJSON) + `'>` + specJSON + `</script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// ServeSwaggerUI serves the traditional Swagger UI (fallback)
func (h *DocsHandler) ServeSwaggerUI() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
}

// ServeSwaggerJSON serves the raw swagger.json file
func (h *DocsHandler) ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	// The swagger.json is served by the httpSwagger handler automatically
	// This is just a custom handler if needed for modifications
	http.Redirect(w, r, "/swagger/doc.json", http.StatusMovedPermanently)
}

// Redirect handles the root docs path
func (h *DocsHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Check if user prefers Scalar or Swagger UI via query param
	if strings.Contains(r.URL.Query().Get("ui"), "swagger") {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
		return
	}
	// Default to Scalar
	http.Redirect(w, r, "/docs/scalar", http.StatusFound)
}
