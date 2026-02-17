package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

type Handlers struct {
	familyStore *store.FamilyStore
}

func NewHandlers(familyStore *store.FamilyStore) *Handlers {
	return &Handlers{familyStore: familyStore}
}

func (h *Handlers) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated"})
		return
	}

	tz, err := h.familyStore.GetTimezone(familyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"timezone": tz})
}

func (h *Handlers) HandleUpdateTimezone(w http.ResponseWriter, r *http.Request) {
	familyID := middleware.GetFamilyID(r)
	if familyID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No family associated"})
		return
	}

	var req struct {
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Timezone == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid timezone",
			"message": "Timezone is required",
		})
		return
	}

	// Validate against IANA timezone database
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid timezone",
			"message": fmt.Sprintf("%q is not a valid IANA timezone identifier", req.Timezone),
		})
		return
	}

	if err := h.familyStore.UpdateTimezone(familyID, req.Timezone); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "Timezone updated",
		"timezone": req.Timezone,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
