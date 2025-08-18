package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type SaveOptionsRequest struct {
	UserID  string      `json:"userId"`
	Options interface{} `json:"options"`
}

type SaveOptionsResponse struct {
	Message string `json:"message"`
}

// In-memory storage for demo (in production, use a database)
var optionsStore = make(map[string]interface{})

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

	// Store options in memory (in production, use proper database)
	optionsStore[req.UserID] = req.Options
	log.Printf("Saved options for user: %s, options: %+v\n", req.UserID, req.Options)

	response := SaveOptionsResponse{
		Message: "Options saved successfully to Go server (Vercel)",
	}

	json.NewEncoder(w).Encode(response)
}
