package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bank-of-dad/internal/config"
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
	_ = store.NewAuthEventStore(db)

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /api/health", handleHealth)

	// Apply middleware chain: CORS → Logging → Routes
	corsMiddleware := middleware.CORS(cfg.FrontendURL, true)
	handler := corsMiddleware(middleware.RequestLogging(mux))

	// Store references for later route registration
	_ = sessionStore // Will be used by auth routes in subsequent phases

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
