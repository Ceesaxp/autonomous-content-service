package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// AuthContext key for storing auth info in request context
type AuthContextKey string

const (
	AuthUserIDKey    AuthContextKey = "user_id"
	AuthUserRoleKey  AuthContextKey = "user_role"
	AuthServiceIDKey AuthContextKey = "service_id"
)

// AuthMiddleware provides authentication middleware for microservices
type AuthMiddleware struct {
	jwtSecret     string
	skipPaths     map[string]bool
	serviceToServiceSecret string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtSecret, serviceSecret string) *AuthMiddleware {
	skipPaths := map[string]bool{
		"/health":  true,
		"/metrics": true,
		"/ready":   true,
		"/live":    true,
	}

	return &AuthMiddleware{
		jwtSecret:              jwtSecret,
		serviceToServiceSecret: serviceSecret,
		skipPaths:              skipPaths,
	}
}

// AddSkipPath adds a path that should skip authentication
func (a *AuthMiddleware) AddSkipPath(path string) {
	a.skipPaths[path] = true
}

// HTTPMiddleware returns HTTP middleware for JWT authentication
func (a *AuthMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain paths
		if a.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT token
		claims, err := a.validateJWTToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := r.Context()
		if userID, ok := claims["user_id"].(string); ok {
			ctx = context.WithValue(ctx, AuthUserIDKey, userID)
		}
		if userRole, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, AuthUserRoleKey, userRole)
		}
		if serviceID, ok := claims["service_id"].(string); ok {
			ctx = context.WithValue(ctx, AuthServiceIDKey, serviceID)
		}

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ServiceToServiceMiddleware provides authentication for service-to-service communication
func (a *AuthMiddleware) ServiceToServiceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks
		if a.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Check for service-to-service authentication
		serviceToken := r.Header.Get("X-Service-Token")
		if serviceToken != "" {
			if serviceToken == a.serviceToServiceSecret {
				// Add service context
				ctx := context.WithValue(r.Context(), AuthServiceIDKey, "internal-service")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "Invalid service token", http.StatusUnauthorized)
			return
		}

		// Fall back to JWT authentication
		a.HTTPMiddleware(next).ServeHTTP(w, r)
	})
}

// validateJWTToken validates a JWT token and returns claims
func (a *AuthMiddleware) validateJWTToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	return claims, nil
}

// RequireRole creates middleware that requires a specific role
func (a *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(AuthUserRoleKey).(string)
			if !ok || role != requiredRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates middleware that requires any of the specified roles
func (a *AuthMiddleware) RequireAnyRole(allowedRoles ...string) func(http.Handler) http.Handler {
	roleMap := make(map[string]bool)
	for _, role := range allowedRoles {
		roleMap[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(AuthUserRoleKey).(string)
			if !ok || !roleMap[role] {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(AuthUserIDKey).(string)
	return userID, ok
}

// GetUserRole extracts user role from request context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(AuthUserRoleKey).(string)
	return role, ok
}

// GetServiceID extracts service ID from request context
func GetServiceID(ctx context.Context) (string, bool) {
	serviceID, ok := ctx.Value(AuthServiceIDKey).(string)
	return serviceID, ok
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"role,omitempty"`
	ServiceID string `json:"service_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenGenerator provides JWT token generation functionality
type TokenGenerator struct {
	secret string
	issuer string
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(secret, issuer string) *TokenGenerator {
	return &TokenGenerator{
		secret: secret,
		issuer: issuer,
	}
}

// GenerateUserToken generates a JWT token for a user
func (tg *TokenGenerator) GenerateUserToken(userID, role string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tg.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tg.secret))
}

// GenerateServiceToken generates a JWT token for service-to-service communication
func (tg *TokenGenerator) GenerateServiceToken(serviceID string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		ServiceID: serviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tg.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tg.secret))
}