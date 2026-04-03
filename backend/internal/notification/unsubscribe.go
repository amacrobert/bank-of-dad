package notification

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"bank-of-dad/repositories"
)

// UnsubscribePayload is the JSON payload embedded in the HMAC token.
type UnsubscribePayload struct {
	ParentID int64 `json:"pid"`
}

// GenerateUnsubscribeToken creates an HMAC-signed token containing the parent ID.
func GenerateUnsubscribeToken(parentID int64, secret []byte) (string, error) {
	payload, err := json.Marshal(UnsubscribePayload{ParentID: parentID})
	if err != nil {
		return "", fmt.Errorf("marshal unsubscribe payload: %w", err)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	// Token format: base64(payload).base64(signature)
	token := base64.URLEncoding.EncodeToString(payload) + "." + base64.URLEncoding.EncodeToString(sig)
	return token, nil
}

// ValidateUnsubscribeToken validates an HMAC-signed token and returns the parent ID.
func ValidateUnsubscribeToken(token string, secret []byte) (int64, error) {
	// Split token into payload and signature
	var payloadB64, sigB64 string
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			payloadB64 = token[:i]
			sigB64 = token[i+1:]
			break
		}
	}
	if payloadB64 == "" || sigB64 == "" {
		return 0, fmt.Errorf("invalid token format")
	}

	payload, err := base64.URLEncoding.DecodeString(payloadB64)
	if err != nil {
		return 0, fmt.Errorf("invalid token payload: %w", err)
	}

	sig, err := base64.URLEncoding.DecodeString(sigB64)
	if err != nil {
		return 0, fmt.Errorf("invalid token signature: %w", err)
	}

	// Verify HMAC
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return 0, fmt.Errorf("invalid token signature")
	}

	var p UnsubscribePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return 0, fmt.Errorf("invalid token payload: %w", err)
	}

	return p.ParentID, nil
}

// UnsubscribeHandler handles GET /api/notifications/unsubscribe
type UnsubscribeHandler struct {
	parentRepo *repositories.ParentRepo
	secret     []byte
}

// NewUnsubscribeHandler creates a new UnsubscribeHandler.
func NewUnsubscribeHandler(parentRepo *repositories.ParentRepo, secret []byte) *UnsubscribeHandler {
	return &UnsubscribeHandler{parentRepo: parentRepo, secret: secret}
}

// HandleUnsubscribe processes the one-click unsubscribe link.
func (h *UnsubscribeHandler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "Invalid or expired unsubscribe link.")
		return
	}

	parentID, err := ValidateUnsubscribeToken(token, h.secret)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "Invalid or expired unsubscribe link.")
		return
	}

	// Set all notification preferences to false
	prefs := map[string]bool{
		"notify_withdrawal_requests": false,
		"notify_chore_completions":   false,
		"notify_decisions":           false,
	}
	if err := h.parentRepo.UpdateNotificationPrefs(parentID, prefs); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Something went wrong. Please try again later.")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "You have been unsubscribed from all Bank of Dad email notifications.\nYou can re-enable notifications in your settings at any time.")
}
