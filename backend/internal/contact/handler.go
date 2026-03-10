package contact

import (
	"bank-of-dad/internal/store"
	"encoding/json"
	"net/http"

	"bank-of-dad/internal/middleware"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type Handler struct {
	brevoClient    *brevo.APIClient
	recipientEmail string
	recipientName  string
	parentStore    *store.ParentStore
}

func NewHandler(
	brevoClient *brevo.APIClient,
	recipientEmail,
	recipientName string,
	parentStore *store.ParentStore) *Handler {
	return &Handler{
		brevoClient,
		recipientEmail,
		recipientName,
		parentStore,
	}
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type ContactSubmissionRequest struct {
	Body string `json:"body"`
}

func (h *Handler) HandleContactSubmission(w http.ResponseWriter, r *http.Request) {
	var submission ContactSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{"invalid_request", "Invalid request body."})
		return
	}

	if submission.Body == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{"invalid_request", "body is required."})
		return
	}

	parentID := middleware.GetUserID(r)
	parent, _ := h.parentStore.GetByID(parentID)

	// TODO: Replace with interface
	_, _, err := h.brevoClient.TransactionalEmailsApi.SendTransacEmail(r.Context(), brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: "noreply@bankofdad.xyz",
			Name:  "Bank of Dad",
		},
		To: []brevo.SendSmtpEmailTo{{
			Email: h.recipientEmail,
			Name:  h.recipientName,
		}},
		ReplyTo: &brevo.SendSmtpEmailReplyTo{
			Email: parent.Email,
			Name:  parent.DisplayName,
		},
		Subject:     "Bank of Dad Contact Submission",
		TextContent: submission.Body,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{"email_error", err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}
