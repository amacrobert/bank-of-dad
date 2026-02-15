package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bank-of-dad/internal/store"
)

type ChildAuth struct {
	familyStore       *store.FamilyStore
	childStore        *store.ChildStore
	refreshTokenStore *store.RefreshTokenStore
	eventStore        *store.AuthEventStore
	jwtKey            []byte
}

func NewChildAuth(
	familyStore *store.FamilyStore,
	childStore *store.ChildStore,
	refreshTokenStore *store.RefreshTokenStore,
	eventStore *store.AuthEventStore,
	jwtKey []byte,
) *ChildAuth {
	return &ChildAuth{
		familyStore:       familyStore,
		childStore:        childStore,
		refreshTokenStore: refreshTokenStore,
		eventStore:        eventStore,
		jwtKey:            jwtKey,
	}
}

func (ca *ChildAuth) HandleChildLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FamilySlug string `json:"family_slug"`
		FirstName  string `json:"first_name"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Find family
	fam, err := ca.familyStore.GetBySlug(req.FamilySlug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if fam == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Family not found"})
		return
	}

	// Find child
	child, err := ca.childStore.GetByFamilyAndName(fam.ID, req.FirstName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if child == nil {
		ca.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
			EventType: "login_failure",
			UserType:  "child",
			FamilyID:  fam.ID,
			IPAddress: clientIP(r),
			Details:   fmt.Sprintf("unknown child name: %s", req.FirstName),
			CreatedAt: time.Now().UTC(),
		})
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid credentials",
			"message": "Hmm, that didn't work. Try again or ask your parent for help!",
		})
		return
	}

	// Check locked
	if child.IsLocked {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"error":   "Account locked",
			"message": "Your account is locked. Ask your parent to help you reset your password.",
		})
		return
	}

	// Check password
	if !ca.childStore.CheckPassword(child, req.Password) {
		attempts, _ := ca.childStore.IncrementFailedAttempts(child.ID)

		ca.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
			EventType: "login_failure",
			UserType:  "child",
			UserID:    child.ID,
			FamilyID:  fam.ID,
			IPAddress: clientIP(r),
			Details:   fmt.Sprintf("wrong password, attempt %d", attempts),
			CreatedAt: time.Now().UTC(),
		})

		if attempts >= 5 {
			ca.childStore.LockAccount(child.ID)       //nolint:errcheck // best-effort cleanup
			ca.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
				EventType: "account_locked",
				UserType:  "child",
				UserID:    child.ID,
				FamilyID:  fam.ID,
				IPAddress: clientIP(r),
				Details:   "locked after 5 failed attempts",
				CreatedAt: time.Now().UTC(),
			})
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"error":   "Account locked",
				"message": "Your account is locked. Ask your parent to help you reset your password.",
			})
			return
		}

		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid credentials",
			"message": "Hmm, that didn't work. Try again or ask your parent for help!",
		})
		return
	}

	// Password correct — reset failed attempts
	ca.childStore.ResetFailedAttempts(child.ID) //nolint:errcheck // best-effort cleanup

	// Generate JWT access token + refresh token (24-hour TTL for children)
	accessToken, err := GenerateAccessToken(ca.jwtKey, "child", child.ID, fam.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create token"})
		return
	}

	refreshToken, err := ca.refreshTokenStore.Create("child", child.ID, fam.ID, 24*time.Hour)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	ca.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
		EventType: "login_success",
		UserType:  "child",
		UserID:    child.ID,
		FamilyID:  fam.ID,
		IPAddress: clientIP(r),
		CreatedAt: time.Now().UTC(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"user_type":   "child",
			"user_id":     child.ID,
			"family_id":   fam.ID,
			"first_name":  child.FirstName,
			"family_slug": req.FamilySlug,
		},
	})
}
