package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"playground/config"
)

type SaveOptionsRequest struct {
	UserID  string      `json:"userId"`
	Options interface{} `json:"options"`
}

type SaveOptionsResponse struct {
	Message string `json:"message"`
}

func GoSaveOptions(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req SaveOptionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	if req.UserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "User ID is required"})
		return
	}

	if req.Options == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Options are required"})
		return
	}

	// Initialize Redis config store - matching TypeScript implementation
	configStore, err := config.NewKVConfigStoreFromEnv()
	if err != nil {
		log.Printf("Failed to initialize config store: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error", "error": err.Error()})
		return
	}
	defer configStore.Close()

	// Store options in Redis with 30-minute expiration (matching TypeScript: ex: 1800)
	ctx := context.Background()
	optionsJSON, err := json.Marshal(req.Options)
	if err != nil {
		log.Printf("Failed to marshal options: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error", "error": "Failed to serialize options"})
		return
	}

	// Use Redis SET with expiration (1800 seconds = 30 minutes, matching TypeScript)
	err = configStore.SetWithExpiration(ctx, req.UserID, string(optionsJSON), 30*time.Minute)
	if err != nil {
		log.Printf("Failed to save options to Redis: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error", "error": "Failed to save options"})
		return
	}

	log.Printf("Saved options for user: %s, options: %+v\n", req.UserID, req.Options)

	response := SaveOptionsResponse{
		Message: "Options saved successfully",
	}

	json.NewEncoder(w).Encode(response)
}
