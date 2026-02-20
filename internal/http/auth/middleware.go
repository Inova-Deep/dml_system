package auth

import (
	"context"
	"net/http"
	"strings"

	logic "github.com/INOVA/DML/internal/logic/auth"
	"github.com/INOVA/DML/internal/response"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	TenantIDKey contextKey = "tenantID"
	RolesKey    contextKey = "roles"
)

// Config dependencies for the middleware
type MiddlewareConfig struct {
	JWTSecret string
}

func AuthMiddleware(cfg MiddlewareConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, "Invalid authorization format. Expected 'Bearer <token>'")
				return
			}

			tokenString := parts[1]
			claims := &logic.Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Extract User ID
			parsedUser, err := uuid.Parse(claims.UserID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Invalid token subject format")
				return
			}
			var pgUserID pgtype.UUID
			pgUserID.Bytes = parsedUser
			pgUserID.Valid = true

			// Extract Tenant ID
			parsedTenant, err := uuid.Parse(claims.TenantID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Invalid token tenant format")
				return
			}
			var pgTenantID pgtype.UUID
			pgTenantID.Bytes = parsedTenant
			pgTenantID.Valid = true

			// Load into request Context
			ctx := context.WithValue(r.Context(), UserIDKey, pgUserID)
			ctx = context.WithValue(ctx, TenantIDKey, pgTenantID)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper functions for handlers to pull context
func GetTenantIDFromContext(ctx context.Context) (pgtype.UUID, bool) {
	val, ok := ctx.Value(TenantIDKey).(pgtype.UUID)
	return val, ok
}

func GetUserIDFromContext(ctx context.Context) (pgtype.UUID, bool) {
	val, ok := ctx.Value(UserIDKey).(pgtype.UUID)
	return val, ok
}

func GetRolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesKey).([]string)
	return val, ok
}

// RequireRole checks if the authenticated user has at least one of the provided roles.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := GetRolesFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			hasAccess := false
			for _, allowed := range allowedRoles {
				for _, ur := range userRoles {
					if ur == allowed {
						hasAccess = true
						break
					}
				}
				if hasAccess {
					break
				}
			}

			if !hasAccess {
				response.Error(w, http.StatusForbidden, "Forbidden: insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
