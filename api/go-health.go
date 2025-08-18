package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func GoHealth(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message:   "Go server (Vercel) is running successfully!",
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}
