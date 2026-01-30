package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	ContextKeyUserType contextKey = "user_type"
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyFamilyID contextKey = "family_id"
)

// SessionValidator looks up a session token and returns user info if valid.
type SessionValidator interface {
	ValidateSession(token string) (userType string, userID int64, familyID int64, err error)
}

func RequireAuth(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			userType, userID, familyID, err := validator.ValidateSession(cookie.Value)
			if err != nil {
				// Clear invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyUserType, userType)
			ctx = context.WithValue(ctx, ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyFamilyID, familyID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireParent wraps RequireAuth and additionally checks that user_type is "parent".
func RequireParent(validator SessionValidator) func(http.Handler) http.Handler {
	authMiddleware := RequireAuth(validator)
	return func(next http.Handler) http.Handler {
		return authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userType, _ := r.Context().Value(ContextKeyUserType).(string)
			if userType != "parent" {
				http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func GetUserType(r *http.Request) string {
	val, _ := r.Context().Value(ContextKeyUserType).(string)
	return val
}

func GetUserID(r *http.Request) int64 {
	val, _ := r.Context().Value(ContextKeyUserID).(int64)
	return val
}

func GetFamilyID(r *http.Request) int64 {
	val, _ := r.Context().Value(ContextKeyFamilyID).(int64)
	return val
}
