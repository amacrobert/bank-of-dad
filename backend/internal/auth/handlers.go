package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

type Handlers struct {
	parentStore  *store.ParentStore
	familyStore  *store.FamilyStore
	childStore   *store.ChildStore
	sessionStore *store.SessionStore
	eventStore   *store.AuthEventStore
	cookieSecure bool
}

func NewHandlers(
	parentStore *store.ParentStore,
	familyStore *store.FamilyStore,
	childStore *store.ChildStore,
	sessionStore *store.SessionStore,
	eventStore *store.AuthEventStore,
	cookieSecure bool,
) *Handlers {
	return &Handlers{
		parentStore:  parentStore,
		familyStore:  familyStore,
		childStore:   childStore,
		sessionStore: sessionStore,
		eventStore:   eventStore,
		cookieSecure: cookieSecure,
	}
}

func (h *Handlers) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userType := middleware.GetUserType(r)
	userID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)

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
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		h.sessionStore.DeleteByToken(cookie.Value) //nolint:errcheck // best-effort cleanup on logout
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	userID := middleware.GetUserID(r)
	familyID := middleware.GetFamilyID(r)
	userType := middleware.GetUserType(r)

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
