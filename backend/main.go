package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bank-of-dad/internal/allowance"
	"bank-of-dad/internal/auth"
	"bank-of-dad/internal/balance"
	"bank-of-dad/internal/config"
	"bank-of-dad/internal/family"
	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	sessionStore := store.NewSessionStore(db)

	// Start session cleanup goroutine (every hour)
	stopCleanup := make(chan struct{})
	defer close(stopCleanup)
	sessionStore.StartCleanupLoop(1*time.Hour, stopCleanup)
	parentStore := store.NewParentStore(db)
	familyStore := store.NewFamilyStore(db)
	childStore := store.NewChildStore(db)
	eventStore := store.NewAuthEventStore(db)

	// Initialize handlers
	googleAuth := auth.NewGoogleAuth(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
		parentStore,
		sessionStore,
		eventStore,
		cfg.FrontendURL,
		cfg.CookieSecure,
	)

	familyHandlers := family.NewHandlers(familyStore, parentStore, childStore, eventStore)
	authHandlers := auth.NewHandlers(parentStore, familyStore, childStore, sessionStore, eventStore, cfg.CookieSecure)
	childAuth := auth.NewChildAuth(familyStore, childStore, sessionStore, eventStore, cfg.CookieSecure)
	txStore := store.NewTransactionStore(db)
	balanceHandler := balance.NewHandler(txStore, childStore)
	scheduleStore := store.NewScheduleStore(db)
	allowanceHandler := allowance.NewHandler(scheduleStore, childStore)

	// Start allowance scheduler goroutine (check every 5 minutes)
	stopScheduler := make(chan struct{})
	defer close(stopScheduler)
	scheduler := allowance.NewScheduler(scheduleStore, txStore, childStore)
	scheduler.Start(5*time.Minute, stopScheduler)

	// Auth middleware
	requireAuth := middleware.RequireAuth(sessionStore)
	requireParent := middleware.RequireParent(sessionStore)

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /api/health", handleHealth)

	// US1: Google OAuth (public)
	mux.HandleFunc("GET /api/auth/google/login", googleAuth.HandleLogin)
	mux.HandleFunc("GET /api/auth/google/callback", googleAuth.HandleCallback)

	// US1: Family management (auth required)
	mux.Handle("POST /api/families", requireParent(http.HandlerFunc(familyHandlers.HandleCreateFamily)))
	mux.Handle("GET /api/families/check-slug", requireParent(http.HandlerFunc(familyHandlers.HandleCheckSlug)))

	// US2: Auth session endpoints (any authenticated user)
	mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(authHandlers.HandleGetMe)))
	mux.Handle("POST /api/auth/logout", requireAuth(http.HandlerFunc(authHandlers.HandleLogout)))

	// US3: Child management (parent auth required)
	mux.Handle("POST /api/children", requireParent(http.HandlerFunc(familyHandlers.HandleCreateChild)))
	mux.Handle("GET /api/children", requireParent(http.HandlerFunc(familyHandlers.HandleListChildren)))

	// US5: Child credential management (parent auth required)
	mux.Handle("PUT /api/children/{id}/password", requireParent(http.HandlerFunc(familyHandlers.HandleResetPassword)))
	mux.Handle("PUT /api/children/{id}/name", requireParent(http.HandlerFunc(familyHandlers.HandleUpdateName)))

	// US4: Child login and family lookup (public)
	childLoginRateLimit := middleware.RateLimit(10, 1*time.Minute)
	mux.Handle("POST /api/auth/child/login", childLoginRateLimit(http.HandlerFunc(childAuth.HandleChildLogin)))
	mux.HandleFunc("GET /api/families/{slug}", familyHandlers.HandleGetFamily)

	// Account Balances (002-account-balances)
	mux.Handle("POST /api/children/{id}/deposit", requireParent(http.HandlerFunc(balanceHandler.HandleDeposit)))
	mux.Handle("POST /api/children/{id}/withdraw", requireParent(http.HandlerFunc(balanceHandler.HandleWithdraw)))
	mux.Handle("GET /api/children/{id}/balance", requireAuth(http.HandlerFunc(balanceHandler.HandleGetBalance)))
	mux.Handle("GET /api/children/{id}/transactions", requireAuth(http.HandlerFunc(balanceHandler.HandleGetTransactions)))

	// Allowance Scheduling (003-allowance-scheduling)
	mux.Handle("POST /api/schedules", requireParent(http.HandlerFunc(allowanceHandler.HandleCreateSchedule)))
	mux.Handle("GET /api/schedules", requireParent(http.HandlerFunc(allowanceHandler.HandleListSchedules)))
	mux.Handle("GET /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleGetSchedule)))
	mux.Handle("PUT /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleUpdateSchedule)))
	mux.Handle("DELETE /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleDeleteSchedule)))
	mux.Handle("POST /api/schedules/{id}/pause", requireParent(http.HandlerFunc(allowanceHandler.HandlePauseSchedule)))
	mux.Handle("POST /api/schedules/{id}/resume", requireParent(http.HandlerFunc(allowanceHandler.HandleResumeSchedule)))
	mux.Handle("GET /api/children/{childId}/upcoming-allowances", requireAuth(http.HandlerFunc(allowanceHandler.HandleGetUpcomingAllowances)))

	// Apply middleware chain: CORS → Logging → Routes
	corsMiddleware := middleware.CORS(cfg.FrontendURL, true)
	handler := corsMiddleware(middleware.RequestLogging(mux))

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Backend server starting on %s...", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
