package family

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

type Handlers struct {
	familyStore *store.FamilyStore
	parentStore *store.ParentStore
	childStore  *store.ChildStore
	eventStore  *store.AuthEventStore
}

func NewHandlers(
	familyStore *store.FamilyStore,
	parentStore *store.ParentStore,
	childStore *store.ChildStore,
	eventStore *store.AuthEventStore,
) *Handlers {
	return &Handlers{
		familyStore: familyStore,
		parentStore: parentStore,
		childStore:  childStore,
		eventStore:  eventStore,
	}
}

func (h *Handlers) HandleCreateFamily(w http.ResponseWriter, r *http.Request) {
	parentID := middleware.GetUserID(r)

	var req struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := ValidateSlug(req.Slug); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid slug format",
			"message": err.Error(),
		})
		return
	}

	exists, err := h.familyStore.SlugExists(req.Slug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if exists {
		suggestions := h.familyStore.SuggestSlugs(req.Slug)
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"error":       "Slug taken",
			"suggestions": suggestions,
		})
		return
	}

	fam, err := h.familyStore.Create(req.Slug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create family"})
		return
	}

	if err := h.parentStore.SetFamilyID(parentID, fam.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to link family"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   fam.ID,
		"slug": fam.Slug,
	})
}

func (h *Handlers) HandleCheckSlug(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slug parameter required"})
		return
	}

	validErr := ValidateSlug(slug)
	valid := validErr == nil

	exists, _ := h.familyStore.SlugExists(slug)

	var suggestions []string
	if exists || !valid {
		suggestions = h.familyStore.SuggestSlugs(slug)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"slug":        slug,
		"available":   !exists && valid,
		"valid":       valid,
		"suggestions": suggestions,
	})
}

func (h *Handlers) HandleGetFamily(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slug required"})
		return
	}

	fam, err := h.familyStore.GetBySlug(slug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"slug":   slug,
		"exists": fam != nil,
	})
}

func (h *Handlers) HandleCreateChild(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated with your account"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := ValidateChildName(req.FirstName); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid name",
			"message": err.Error(),
		})
		return
	}

	if err := ValidateChildPassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Password too short",
			"message": "Password must be at least 6 characters.",
		})
		return
	}

	child, err := h.childStore.Create(familyID, req.FirstName, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"error":   "Name taken",
				"message": "A child named " + req.FirstName + " already exists in your family.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create child"})
		return
	}

	// Get family slug for login URL
	fam, _ := h.familyStore.GetByID(familyID)
	var familySlug string
	if fam != nil {
		familySlug = fam.Slug
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          child.ID,
		"first_name":  child.FirstName,
		"family_slug": familySlug,
		"login_url":   "/" + familySlug,
	})
}

func (h *Handlers) HandleListChildren(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"children": []interface{}{}})
		return
	}

	children, err := h.childStore.ListByFamily(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list children"})
		return
	}

	type childResponse struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		IsLocked  bool   `json:"is_locked"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]childResponse, len(children))
	for i, c := range children {
		result[i] = childResponse{
			ID:        c.ID,
			FirstName: c.FirstName,
			IsLocked:  c.IsLocked,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"children": result})
}

func (h *Handlers) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	childID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid child ID"})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Child not found"})
		return
	}

	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := ValidateChildPassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password too short"})
		return
	}

	wasLocked := child.IsLocked
	if err := h.childStore.UpdatePassword(childID, req.Password); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
		return
	}

	h.eventStore.LogEvent(store.AuthEvent{
		EventType: "password_reset",
		UserType:  "parent",
		UserID:    middleware.GetUserID(r),
		FamilyID:  familyID,
		IPAddress: r.RemoteAddr,
		Details:   fmt.Sprintf("reset password for child %d (%s)", childID, child.FirstName),
		CreatedAt: time.Now().UTC(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Password updated",
		"account_unlocked": wasLocked,
	})
}

func (h *Handlers) HandleUpdateName(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	childID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid child ID"})
		return
	}

	child, err := h.childStore.GetByID(childID)
	if err != nil || child == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Child not found"})
		return
	}

	if child.FamilyID != familyID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := ValidateChildName(req.FirstName); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid name",
			"message": err.Error(),
		})
		return
	}

	if err := h.childStore.UpdateName(childID, familyID, req.FirstName); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Name taken"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update name"})
		return
	}

	h.eventStore.LogEvent(store.AuthEvent{
		EventType: "name_updated",
		UserType:  "parent",
		UserID:    middleware.GetUserID(r),
		FamilyID:  familyID,
		IPAddress: r.RemoteAddr,
		Details:   fmt.Sprintf("updated child %d name from %q to %q", childID, child.FirstName, req.FirstName),
		CreatedAt: time.Now().UTC(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Name updated",
		"first_name": req.FirstName,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
