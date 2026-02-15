package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bank-of-dad/internal/store"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuth struct {
	config            *oauth2.Config
	parentStore       *store.ParentStore
	refreshTokenStore *store.RefreshTokenStore
	eventStore        *store.AuthEventStore
	frontendURL       string
	jwtKey            []byte
}

func NewGoogleAuth(
	clientID, clientSecret, redirectURL string,
	parentStore *store.ParentStore,
	refreshTokenStore *store.RefreshTokenStore,
	eventStore *store.AuthEventStore,
	frontendURL string,
	jwtKey []byte,
) *GoogleAuth {
	return &GoogleAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		parentStore:       parentStore,
		refreshTokenStore: refreshTokenStore,
		eventStore:        eventStore,
		frontendURL:       frontendURL,
		jwtKey:            jwtKey,
	}
}

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (g *GoogleAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   strings.HasPrefix(g.config.RedirectURL, "https://"),
		SameSite: http.SameSiteLaxMode,
	})

	url := g.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (g *GoogleAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		http.Error(w, `{"error":"Missing state parameter"}`, http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		http.Error(w, `{"error":"Invalid state parameter"}`, http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := g.config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Google token exchange failed: %v", err)
		http.Error(w, `{"error":"Authentication failed"}`, http.StatusUnauthorized)
		return
	}

	// Fetch user info
	client := g.config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Failed to fetch Google userinfo: %v", err)
		http.Error(w, `{"error":"Failed to get user info"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on HTTP response body

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, `{"error":"Failed to parse user info"}`, http.StatusInternalServerError)
		return
	}

	// Find or create parent
	parent, err := g.parentStore.GetByGoogleID(userInfo.ID)
	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	isNewUser := parent == nil
	if isNewUser {
		parent, err = g.parentStore.Create(userInfo.ID, userInfo.Email, userInfo.Name)
		if err != nil {
			http.Error(w, `{"error":"Failed to create account"}`, http.StatusInternalServerError)
			return
		}

		g.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
			EventType: "account_created",
			UserType:  "parent",
			UserID:    parent.ID,
			IPAddress: clientIP(r),
			Details:   fmt.Sprintf("registered via Google: %s", userInfo.Email),
			CreatedAt: time.Now().UTC(),
		})
	}

	// Generate JWT access token + refresh token
	accessToken, err := GenerateAccessToken(g.jwtKey, "parent", parent.ID, parent.FamilyID)
	if err != nil {
		http.Error(w, `{"error":"Failed to create token"}`, http.StatusInternalServerError)
		return
	}

	refreshToken, err := g.refreshTokenStore.Create("parent", parent.ID, parent.FamilyID, 7*24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	g.eventStore.LogEvent(store.AuthEvent{ //nolint:errcheck // best-effort audit logging
		EventType: "login_success",
		UserType:  "parent",
		UserID:    parent.ID,
		FamilyID:  parent.FamilyID,
		IPAddress: clientIP(r),
		CreatedAt: time.Now().UTC(),
	})

	// Redirect to frontend callback with tokens in URL params
	redirect := "/dashboard"
	if isNewUser || parent.FamilyID == 0 {
		redirect = "/setup"
	}

	callbackURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&redirect=%s",
		g.frontendURL,
		url.QueryEscape(accessToken),
		url.QueryEscape(refreshToken),
		url.QueryEscape(redirect),
	)
	http.Redirect(w, r, callbackURL, http.StatusFound)
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
