package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type MessageResponse struct {
	Message string `json:"message"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	response := MessageResponse{
		Message: "Hello from Bank of Dad!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/message", messageHandler)

	handler := corsMiddleware(mux)

	log.Println("Backend server starting on port 8001...")
	if err := http.ListenAndServe(":8001", handler); err != nil {
		log.Fatal(err)
	}
}
