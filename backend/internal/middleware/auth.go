package middleware

import (
	"context"
	"net/http"
	"strings"

	"bank-of-dad/internal/auth"
)

func RequireAuth(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateAccessToken(jwtKey, parts[1])
			if err != nil {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, auth.ContextKeyUserType, claims.UserType)
			ctx = context.WithValue(ctx, auth.ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, auth.ContextKeyFamilyID, claims.FamilyID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Context key aliases — canonical definitions live in auth package.
var (
	ContextKeyUserType = auth.ContextKeyUserType
	ContextKeyUserID   = auth.ContextKeyUserID
	ContextKeyFamilyID = auth.ContextKeyFamilyID
)

// Context helper wrappers — delegate to auth package to avoid import cycles.
func GetUserType(r *http.Request) string  { return auth.GetUserType(r) }
func GetUserID(r *http.Request) int64     { return auth.GetUserID(r) }
func GetFamilyID(r *http.Request) int64   { return auth.GetFamilyID(r) }

// RequireParent wraps RequireAuth and additionally checks that user_type is "parent".
func RequireParent(jwtKey []byte) func(http.Handler) http.Handler {
	authMiddleware := RequireAuth(jwtKey)
	return func(next http.Handler) http.Handler {
		return authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userType := auth.GetUserType(r)
			if userType != "parent" {
				http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
