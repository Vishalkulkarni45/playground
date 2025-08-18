package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Response structure for JSON responses
type Response struct {
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// SaveOptions request structure
type SaveOptionsRequest struct {
	UserID  string      `json:"userId"`
	Options interface{} `json:"options"`
}

// SaveOptions response structure
type SaveOptionsResponse struct {
	Message string `json:"message"`
}

// Verify request structure (basic for testing)
type VerifyRequest struct {
	AttestationID   string      `json:"attestationId"`
	Proof           interface{} `json:"proof"`
	PublicSignals   interface{} `json:"publicSignals"`
	UserContextData interface{} `json:"userContextData"`
}

// Verify response structure (basic for testing)
type VerifyResponse struct {
	Status              string      `json:"status"`
	Result              bool        `json:"result"`
	Message             string      `json:"message,omitempty"`
	CredentialSubject   interface{} `json:"credentialSubject,omitempty"`
	VerificationOptions interface{} `json:"verificationOptions,omitempty"`
}

// In-memory storage for testing (replace with proper database later)
var optionsStore = make(map[string]interface{})

// CORS middleware
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message:   "Go server is running successfully!",
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// SaveOptions endpoint - mimics the TypeScript saveOptions.ts functionality
func saveOptionsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
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
	fmt.Printf("Saved options for user: %s, options: %+v\n", req.UserID, req.Options)

	response := SaveOptionsResponse{
		Message: "Options saved successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Verify endpoint - basic structure for testing (will implement Self verification later)
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON"})
		return
	}

	// Basic validation
	if req.AttestationID == "" || req.Proof == nil || req.PublicSignals == nil || req.UserContextData == nil {
		w.WriteHeader(http.StatusBadRequest)
		response := VerifyResponse{
			Status:  "error",
			Result:  false,
			Message: "Proof, publicSignals, attestationId and userContextData are required",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	fmt.Printf("Received verification request for attestationId: %s\n", req.AttestationID)

	// For now, return a mock successful response for testing
	// TODO: Implement actual Self protocol verification
	response := VerifyResponse{
		Status:  "success",
		Result:  true,
		Message: "Mock verification successful (Go server working!)",
		CredentialSubject: map[string]interface{}{
			"nationality":  "Test Country",
			"issuingState": "Test State",
			"name":         "Test User",
			"dateOfBirth":  "1990-01-01",
			"idNumber":     "TEST123",
			"gender":       "Test",
			"expiryDate":   "2030-01-01",
		},
		VerificationOptions: map[string]interface{}{
			"minimumAge":        18,
			"ofac":              true,
			"excludedCountries": []string{},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Hello World endpoint
func helloHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message:   "Hello, World! This is your Go server.",
		Status:    "success",
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "text/html")

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Go Server</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
			.container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #333; text-align: center; }
			.endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
			.method { color: #28a745; font-weight: bold; }
			a { color: #007bff; text-decoration: none; }
			a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>🚀 Go Server is Running!</h1>
			<p>Welcome to your Go server. Here are the available endpoints:</p>
			
			<div class="endpoint">
				<strong class="method">GET</strong> <a href="/health">/health</a>
				<p>Health check endpoint that returns server status</p>
			</div>
			
			<div class="endpoint">
				<strong class="method">GET</strong> <a href="/hello">/hello</a>
				<p>Simple Hello World endpoint</p>
			</div>
			
			<div class="endpoint">
				<strong class="method">POST</strong> /api/saveOptions
				<p>Save verification options (mimics TypeScript version)</p>
			</div>
			
			<div class="endpoint">
				<strong class="method">POST</strong> /api/verify
				<p>Verify passport proof (basic mock for testing)</p>
			</div>
			
			<div class="endpoint">
				<strong class="method">GET</strong> <a href="/">/</a>
				<p>This page - server information</p>
			</div>
			
			<p style="text-align: center; color: #666; margin-top: 30px;">
				Server started at: ` + time.Now().Format("2006-01-02 15:04:05") + `
			</p>
		</div>
	</body>
	</html>`

	fmt.Fprint(w, html)
}

func main() {
	// Set up routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/api/saveOptions", saveOptionsHandler)
	http.HandleFunc("/api/verify", verifyHandler)

	// Server configuration
	port := "8080"

	fmt.Printf("🚀 Starting Go server on port %s...\n", port)
	fmt.Printf("📍 Server will be available at: http://localhost:%s\n", port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   GET / - Server information page")
	fmt.Println("   GET /health - Health check")
	fmt.Println("   GET /hello - Hello World")
	fmt.Println("   POST /api/saveOptions - Save verification options")
	fmt.Println("   POST /api/verify - Verify passport proof (mock)")
	fmt.Println("\n✨ Press Ctrl+C to stop the server")
	fmt.Println("🔗 CORS enabled for frontend integration")

	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
