package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"bank-of-dad/internal/store"
)

type Handlers struct {
	parentStore       *store.ParentStore
	familyStore       *store.FamilyStore
	childStore        *store.ChildStore
	refreshTokenStore *store.RefreshTokenStore
	eventStore        *store.AuthEventStore
	jwtKey            []byte
}

func NewHandlers(
	parentStore *store.ParentStore,
	familyStore *store.FamilyStore,
	childStore *store.ChildStore,
	refreshTokenStore *store.RefreshTokenStore,
	eventStore *store.AuthEventStore,
	jwtKey []byte,
) *Handlers {
	return &Handlers{
		parentStore:       parentStore,
		familyStore:       familyStore,
		childStore:        childStore,
		refreshTokenStore: refreshTokenStore,
		eventStore:        eventStore,
		jwtKey:            jwtKey,
	}
}

func (h *Handlers) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userType := GetUserType(r)
	userID := GetUserID(r)
	familyID := GetFamilyID(r)

	var familySlug string
	if familyID != 0 {
		fam, err := h.familyStore.GetByID(familyID)
		if err == nil && fam != nil {
			familySlug = fam.Slug
		}
	}

	if userType == "parent" {
		parent, err := h.parentStore.GetByID(userID)
		if err != nil || parent == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"user_type":    "parent",
			"user_id":      parent.ID,
			"family_id":    parent.FamilyID,
			"display_name": parent.DisplayName,
			"email":        parent.Email,
			"family_slug":  familySlug,
		})
		return
	}

	// Child user
	child, err := h.childStore.GetByID(userID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_type":   "child",
		"user_id":     child.ID,
		"family_id":   familyID,
		"first_name":  child.FirstName,
		"family_slug": familySlug,
		"avatar":      child.Avatar,
	})
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
		tokenHash := store.HashToken(req.RefreshToken)
		h.refreshTokenStore.DeleteByHash(tokenHash) //nolint:errcheck // best-effort cleanup on logout
	}

	userID := GetUserID(r)
	familyID := GetFamilyID(r)
	userType := GetUserType(r)

	h.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
		EventType: "logout",
		UserType:  userType,
		UserID:    userID,
		FamilyID:  familyID,
		IPAddress: clientIP(r),
		CreatedAt: time.Now().UTC(),
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

func (h *Handlers) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	// Validate the refresh token
	rt, err := h.refreshTokenStore.Validate(req.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if rt == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired refresh token"})
		return
	}

	// Delete old refresh token (rotation)
	h.refreshTokenStore.DeleteByHash(rt.TokenHash) //nolint:errcheck // best-effort deletion

	// Look up current user data to get current family_id
	var familyID int64
	switch rt.UserType {
	case "parent":
		parent, err := h.parentStore.GetByID(rt.UserID)
		if err != nil || parent == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "User not found"})
			return
		}
		familyID = parent.FamilyID
	case "child":
		child, err := h.childStore.GetByID(rt.UserID)
		if err != nil || child == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "User not found"})
			return
		}
		familyID = child.FamilyID
	default:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid user type"})
		return
	}

	// Determine TTL based on user type
	ttl := 7 * 24 * time.Hour // parent default
	if rt.UserType == "child" {
		ttl = 24 * time.Hour
	}

	// Create new refresh token
	newRefreshToken, err := h.refreshTokenStore.Create(rt.UserType, rt.UserID, familyID, ttl)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	// Generate new access token
	accessToken, err := GenerateAccessToken(h.jwtKey, rt.UserType, rt.UserID, familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
