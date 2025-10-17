package auth

import (
	"boilerplate/internal/config"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// UserContextKey is the key used to store user claims in request context
	UserContextKey contextKey = "user"
)

// UserClaims represents the claims extracted from JWT
type UserClaims struct {
	Subject  string   `json:"sub"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	ClientID string   `json:"azp"`
}

// Middleware provides JWT authentication middleware
type Middleware struct {
	config    config.AuthConfig
	logger    *slog.Logger
	jwksCache *jwksCache
}

func NewMiddleware(cfg config.AuthConfig, logger *slog.Logger) *Middleware {
	m := &Middleware{
		config: cfg,
		logger: logger,
		jwksCache: &jwksCache{
			keys: make(map[string]*rsa.PublicKey),
		},
	}

	// Pre-load JWKS if configured
	if cfg.JWKSURL != "" {
		go m.refreshJWKS()
	}

	return m
}

// Authenticate is the HTTP middleware that validates JWT tokens
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If auth is not enabled, skip validation
		if !m.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Debug("missing authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.logger.Debug("invalid authorization header format")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.logger.Debug("token validation failed", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a JWT token and returns the claims
func (m *Middleware) validateToken(tokenString string) (*UserClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		publicKey, err := m.jwksCache.getKey(kid)
		if err != nil {
			if err := m.refreshJWKS(); err != nil {
				return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
			}
			publicKey, err = m.jwksCache.getKey(kid)
			if err != nil {
				return nil, fmt.Errorf("key not found in JWKS: %w", err)
			}
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract standard claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// Validate issuer
	if m.config.Issuer != "" {
		iss, ok := claims["iss"].(string)
		if !ok || iss != m.config.Issuer {
			return nil, errors.New("invalid issuer")
		}
	}

	// Validate expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("missing exp claim")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, errors.New("token expired")
	}

	// Extract user claims
	userClaims := &UserClaims{}

	if sub, ok := claims["sub"].(string); ok {
		userClaims.Subject = sub
	}

	if email, ok := claims["email"].(string); ok {
		userClaims.Email = email
	}

	if name, ok := claims["name"].(string); ok {
		userClaims.Name = name
	}

	if azp, ok := claims["azp"].(string); ok {
		userClaims.ClientID = azp
	}

	// Extract roles (can be in different claim names depending on provider)
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles = append(userClaims.Roles, roleStr)
				}
			}
		}
	}

	return userClaims, nil
}

// refreshJWKS fetches the latest JWKS from the provider
func (m *Middleware) refreshJWKS() error {
	if m.config.JWKSURL == "" {
		return errors.New("JWKS URL not configured")
	}

	resp, err := http.Get(m.config.JWKSURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Parse and cache public keys
	m.jwksCache.mu.Lock()
	defer m.jwksCache.mu.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}

		publicKey, err := parseRSAPublicKey(key.N, key.E)
		if err != nil {
			m.logger.Warn("failed to parse public key", "kid", key.Kid, "error", err)
			continue
		}

		m.jwksCache.keys[key.Kid] = publicKey
	}

	m.logger.Info("refreshed JWKS", "key_count", len(m.jwksCache.keys))
	return nil
}

// jwksCache holds cached JWKS public keys
type jwksCache struct {
	mu   sync.RWMutex
	keys map[string]*rsa.PublicKey
}

func (c *jwksCache) getKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key, ok := c.keys[kid]
	if !ok {
		return nil, errors.New("key not found")
	}

	return key, nil
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

func GetUserClaims(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	return claims, ok
}
