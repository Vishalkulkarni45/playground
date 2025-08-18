package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type VerifyRequest struct {
	AttestationID   string      `json:"attestationId"`
	Proof           interface{} `json:"proof"`
	PublicSignals   interface{} `json:"publicSignals"`
	UserContextData interface{} `json:"userContextData"`
}

type VerifyResponse struct {
	Status              string      `json:"status"`
	Result              bool        `json:"result"`
	Message             string      `json:"message,omitempty"`
	CredentialSubject   interface{} `json:"credentialSubject,omitempty"`
	VerificationOptions interface{} `json:"verificationOptions,omitempty"`
}

func GoVerify(w http.ResponseWriter, r *http.Request) {
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

	// Read the raw body first for debugging
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error reading request body"})
		return
	}

	log.Printf("Received request body: %s", string(body))

	// Try to parse as generic map first to see the structure
	var genericReq map[string]interface{}
	if err := json.Unmarshal(body, &genericReq); err != nil {
		log.Printf("JSON parsing error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message":  "Invalid JSON",
			"error":    err.Error(),
			"received": string(body),
		})
		return
	}

	log.Printf("Parsed JSON structure: %+v", genericReq)

	// Now try to parse into our expected structure
	var req VerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Struct parsing error: %v", err)
		// Return success anyway for now to see what we're getting
		response := VerifyResponse{
			Status:  "debug",
			Result:  true,
			Message: fmt.Sprintf("Debug: received data but couldn't parse into expected struct. Raw: %s", string(body)),
		}
		json.NewEncoder(w).Encode(response)
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

	log.Printf("Received verification request for attestationId: %s\n", req.AttestationID)

	// For now, return a mock successful response for testing
	// TODO: Implement actual Self protocol verification
	response := VerifyResponse{
		Status:  "success",
		Result:  true,
		Message: "Mock verification successful (Go server on Vercel working!)",
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
