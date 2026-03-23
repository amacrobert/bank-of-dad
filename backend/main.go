package main

import (
	"bank-of-dad/internal/contact"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	_ "time/tzdata" // Embed IANA timezone database for environments without system tzdata (e.g. Alpine Docker)

	brevo "github.com/getbrevo/brevo-go/lib"

	"bank-of-dad/internal/allowance"
	"bank-of-dad/internal/auth"
	"bank-of-dad/internal/balance"
	"bank-of-dad/internal/chore"
	"bank-of-dad/internal/config"
	"bank-of-dad/internal/family"
	"bank-of-dad/internal/goals"
	"bank-of-dad/internal/interest"
	"bank-of-dad/internal/middleware"
	"bank-of-dad/internal/settings"
	"bank-of-dad/internal/subscription"
	"bank-of-dad/repositories"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repositories.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	defer sqlDB.Close() //nolint:errcheck // deferred close on application shutdown

	refreshTokenRepo := repositories.NewRefreshTokenRepo(db)

	// Start refresh token cleanup goroutine (every hour)
	stopCleanup := make(chan struct{})
	defer close(stopCleanup)
	refreshTokenRepo.StartCleanupLoop(1*time.Hour, stopCleanup)

	jwtKey := cfg.JWTSecret

	parentRepo := repositories.NewParentRepo(db)
	familyRepo := repositories.NewFamilyRepo(db)
	childRepo := repositories.NewChildRepo(db)
	eventRepo := repositories.NewAuthEventRepo(db)

	// Set up Brevo
	brevoConfig := brevo.NewConfiguration()
	brevoConfig.AddDefaultHeader("api-key", cfg.BrevoApiKey)
	brevoClient := brevo.NewAPIClient(brevoConfig)

	// Initialize handlers
	googleAuth := auth.NewGoogleAuth(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
		parentRepo,
		refreshTokenRepo,
		eventRepo,
		cfg.FrontendURL,
		jwtKey,
	)

	familyHandlers := family.NewHandlers(familyRepo, parentRepo, childRepo, eventRepo, jwtKey)
	authHandlers := auth.NewHandlers(parentRepo, familyRepo, childRepo, refreshTokenRepo, eventRepo, jwtKey)
	childAuth := auth.NewChildAuth(familyRepo, childRepo, refreshTokenRepo, eventRepo, jwtKey)
	txRepo := repositories.NewTransactionRepo(db)
	interestRepo := repositories.NewInterestRepo(db)
	interestScheduleRepo := repositories.NewInterestScheduleRepo(db)
	goalRepo := repositories.NewSavingsGoalRepo(db)
	goalAllocationRepo := repositories.NewGoalAllocationRepo(db)
	balanceHandler := balance.NewHandler(txRepo, childRepo, interestRepo, interestScheduleRepo, goalRepo)
	scheduleRepo := repositories.NewScheduleRepo(db)
	allowanceHandler := allowance.NewHandler(scheduleRepo, childRepo, familyRepo)
	interestHandler := interest.NewHandler(interestRepo, childRepo, interestScheduleRepo, familyRepo)
	settingsHandlers := settings.NewHandlers(familyRepo)
	goalsHandler := goals.NewHandler(goalRepo, childRepo, goalAllocationRepo)
	webhookEventRepo := repositories.NewWebhookEventRepo(db)
	subscriptionHandlers := subscription.NewHandlers(familyRepo, parentRepo, childRepo, webhookEventRepo, cfg.StripeSecretKey, cfg.StripeWebhookSecret, cfg.FrontendURL)
	contactHandler := contact.NewHandler(brevoClient, cfg.ContactRecipientEmail, cfg.ContactRecipientName, parentRepo)
	choreRepo := repositories.NewChoreRepo(db)
	choreInstanceRepo := repositories.NewChoreInstanceRepo(db)
	choreHandler := chore.NewHandler(choreRepo, choreInstanceRepo, txRepo, childRepo)

	// Start allowance scheduler goroutine (check every 5 minutes)
	stopAllowanceScheduler := make(chan struct{})
	defer close(stopAllowanceScheduler)
	allowanceScheduler := allowance.NewScheduler(scheduleRepo, txRepo, childRepo)
	allowanceScheduler.Start(5*time.Minute, stopAllowanceScheduler)

	// Start chore scheduler goroutine (check every 5 minutes)
	stopChoreScheduler := make(chan struct{})
	defer close(stopChoreScheduler)
	choreScheduler := chore.NewScheduler(choreRepo, choreInstanceRepo, childRepo, familyRepo)
	choreScheduler.Start(5*time.Minute, stopChoreScheduler)

	// Start interest accrual scheduler goroutine (check every hour)
	stopInterestScheduler := make(chan struct{})
	defer close(stopInterestScheduler)
	interestScheduler := interest.NewScheduler(interestRepo)
	interestScheduler.Start(1*time.Hour, stopInterestScheduler)

	// Auth middleware
	requireAuth := middleware.RequireAuth(jwtKey)
	requireParent := middleware.RequireParent(jwtKey)

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
	mux.Handle("DELETE /api/children/{id}", requireParent(http.HandlerFunc(familyHandlers.HandleDeleteChild)))

	// Token refresh (public — access token may be expired)
	refreshRateLimit := middleware.RateLimit(10, 1*time.Minute)
	mux.Handle("POST /api/auth/refresh", refreshRateLimit(http.HandlerFunc(authHandlers.HandleRefresh)))

	// US4: Child login and family lookup (public)
	childLoginRateLimit := middleware.RateLimit(10, 1*time.Minute)
	mux.Handle("POST /api/auth/child/login", childLoginRateLimit(http.HandlerFunc(childAuth.HandleChildLogin)))
	mux.HandleFunc("GET /api/families/{slug}", familyHandlers.HandleGetFamily)
	familyChildrenRateLimit := middleware.RateLimit(30, 1*time.Minute)
	mux.Handle("GET /api/families/{slug}/children", familyChildrenRateLimit(http.HandlerFunc(familyHandlers.HandleListFamilyChildren)))

	// Account Balances (002-account-balances)
	mux.Handle("POST /api/children/{id}/deposit", requireParent(http.HandlerFunc(balanceHandler.HandleDeposit)))
	mux.Handle("POST /api/children/{id}/withdraw", requireParent(http.HandlerFunc(balanceHandler.HandleWithdraw)))
	mux.Handle("GET /api/children/{id}/balance", requireAuth(http.HandlerFunc(balanceHandler.HandleGetBalance)))
	mux.Handle("GET /api/children/{id}/transactions", requireAuth(http.HandlerFunc(balanceHandler.HandleGetTransactions)))

	// Interest (combined rate + schedule)
	mux.Handle("PUT /api/children/{id}/interest", requireParent(http.HandlerFunc(interestHandler.HandleSetInterest)))

	// Allowance Scheduling (003-allowance-scheduling)
	mux.Handle("POST /api/schedules", requireParent(http.HandlerFunc(allowanceHandler.HandleCreateSchedule)))
	mux.Handle("GET /api/schedules", requireParent(http.HandlerFunc(allowanceHandler.HandleListSchedules)))
	mux.Handle("GET /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleGetSchedule)))
	mux.Handle("PUT /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleUpdateSchedule)))
	mux.Handle("DELETE /api/schedules/{id}", requireParent(http.HandlerFunc(allowanceHandler.HandleDeleteSchedule)))
	mux.Handle("POST /api/schedules/{id}/pause", requireParent(http.HandlerFunc(allowanceHandler.HandlePauseSchedule)))
	mux.Handle("POST /api/schedules/{id}/resume", requireParent(http.HandlerFunc(allowanceHandler.HandleResumeSchedule)))
	mux.Handle("GET /api/children/{childId}/upcoming-allowances", requireAuth(http.HandlerFunc(allowanceHandler.HandleGetUpcomingAllowances)))

	// Interest schedule endpoints (006-account-management-enhancements)
	mux.Handle("GET /api/children/{childId}/interest-schedule", requireAuth(http.HandlerFunc(interestHandler.HandleGetInterestSchedule)))

	// Child settings (017-child-visual-themes, 019-child-self-avatar)
	mux.Handle("PUT /api/child/settings/theme", requireAuth(http.HandlerFunc(familyHandlers.HandleUpdateTheme)))
	mux.Handle("PUT /api/child/settings/avatar", requireAuth(http.HandlerFunc(familyHandlers.HandleUpdateAvatar)))

	// Settings (013-parent-settings)
	mux.Handle("GET /api/settings", requireParent(http.HandlerFunc(settingsHandlers.HandleGetSettings)))
	mux.Handle("PUT /api/settings/timezone", requireParent(http.HandlerFunc(settingsHandlers.HandleUpdateTimezone)))
	mux.Handle("PUT /api/settings/bank-name", requireParent(http.HandlerFunc(settingsHandlers.HandleUpdateBankName)))

	// Subscription (024-stripe-subscription)
	mux.Handle("GET /api/subscription", requireParent(http.HandlerFunc(subscriptionHandlers.HandleGetSubscription)))
	mux.Handle("POST /api/subscription/checkout", requireParent(http.HandlerFunc(subscriptionHandlers.HandleCreateCheckout)))
	mux.Handle("POST /api/subscription/portal", requireParent(http.HandlerFunc(subscriptionHandlers.HandleCreatePortal)))
	mux.HandleFunc("POST /api/stripe/webhook", subscriptionHandlers.HandleStripeWebhook)

	// Savings Goals (025-savings-goals)
	mux.Handle("GET /api/children/{id}/savings-goals", requireAuth(http.HandlerFunc(goalsHandler.HandleList)))
	mux.Handle("POST /api/children/{id}/savings-goals", requireAuth(http.HandlerFunc(goalsHandler.HandleCreate)))
	mux.Handle("PUT /api/children/{id}/savings-goals/{goalId}", requireAuth(http.HandlerFunc(goalsHandler.HandleUpdate)))
	mux.Handle("DELETE /api/children/{id}/savings-goals/{goalId}", requireAuth(http.HandlerFunc(goalsHandler.HandleDelete)))
	mux.Handle("POST /api/children/{id}/savings-goals/{goalId}/allocate", requireAuth(http.HandlerFunc(goalsHandler.HandleAllocate)))
	mux.Handle("GET /api/children/{id}/savings-goals/{goalId}/allocations", requireAuth(http.HandlerFunc(goalsHandler.HandleListAllocations)))

	// Account deletion
	mux.Handle("DELETE /api/account", requireParent(http.HandlerFunc(familyHandlers.HandleDeleteAccount)))

	// Contact form submission
	mux.Handle("POST /api/contact", requireParent(http.HandlerFunc(contactHandler.HandleContactSubmission)))

	// Child-scoped allowance endpoints (006-account-management-enhancements)
	mux.Handle("GET /api/children/{childId}/allowance", requireAuth(http.HandlerFunc(allowanceHandler.HandleGetChildAllowance)))
	mux.Handle("PUT /api/children/{childId}/allowance", requireParent(http.HandlerFunc(allowanceHandler.HandleSetChildAllowance)))
	mux.Handle("DELETE /api/children/{childId}/allowance", requireParent(http.HandlerFunc(allowanceHandler.HandleDeleteChildAllowance)))
	mux.Handle("POST /api/children/{childId}/allowance/pause", requireParent(http.HandlerFunc(allowanceHandler.HandlePauseChildAllowance)))
	mux.Handle("POST /api/children/{childId}/allowance/resume", requireParent(http.HandlerFunc(allowanceHandler.HandleResumeChildAllowance)))

	// Chore System (031-chore-system)
	mux.Handle("POST /api/chores", requireParent(http.HandlerFunc(choreHandler.HandleCreateChore)))
	mux.Handle("GET /api/chores", requireParent(http.HandlerFunc(choreHandler.HandleListChores)))

	// Chore instances — child endpoints
	mux.Handle("GET /api/child/chores", requireAuth(http.HandlerFunc(choreHandler.HandleChildListChores)))
	mux.Handle("POST /api/child/chores/{id}/complete", requireAuth(http.HandlerFunc(choreHandler.HandleCompleteChore)))
	mux.Handle("GET /api/child/chores/earnings", requireAuth(http.HandlerFunc(choreHandler.HandleChildEarnings)))

	// Chore instances — parent endpoints
	mux.Handle("GET /api/chores/pending", requireParent(http.HandlerFunc(choreHandler.HandleListPending)))
	mux.Handle("POST /api/chore-instances/{id}/approve", requireParent(http.HandlerFunc(choreHandler.HandleApprove)))
	mux.Handle("POST /api/chore-instances/{id}/reject", requireParent(http.HandlerFunc(choreHandler.HandleReject)))
	mux.Handle("PUT /api/chores/{id}", requireParent(http.HandlerFunc(choreHandler.HandleUpdateChore)))
	mux.Handle("DELETE /api/chores/{id}", requireParent(http.HandlerFunc(choreHandler.HandleDeleteChore)))
	mux.Handle("PATCH /api/chores/{id}/activate", requireParent(http.HandlerFunc(choreHandler.HandleActivate)))
	mux.Handle("PATCH /api/chores/{id}/deactivate", requireParent(http.HandlerFunc(choreHandler.HandleDeactivate)))

	// Apply middleware chain: CORS → Logging → Routes
	corsMiddleware := middleware.CORS(cfg.FrontendURL)
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
