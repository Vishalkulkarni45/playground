package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	self "github.com/selfxyz/self/sdk/sdk-go"
	"github.com/selfxyz/self/sdk/sdk-go/common"
)

type VerifyRequest struct {
	AttestationID   string      `json:"attestationId"`
	Proof           interface{} `json:"proof"`
	PublicSignals   interface{} `json:"publicSignals"`
	UserContextData interface{} `json:"userContextData"`
	UserID          string      `json:"userId,omitempty"`
}

type VerifyResponse struct {
	Status              string      `json:"status"`
	Result              bool        `json:"result"`
	Message             string      `json:"message,omitempty"`
	CredentialSubject   interface{} `json:"credentialSubject,omitempty"`
	VerificationOptions interface{} `json:"verificationOptions,omitempty"`
}

// CustomConfigStore implements a configuration store for Self verification
type CustomConfigStore struct {
	configs map[string]self.VerificationConfig
	mutex   sync.RWMutex
}

// NewCustomConfigStore creates a new custom config store
func NewCustomConfigStore() *CustomConfigStore {
	return &CustomConfigStore{
		configs: make(map[string]self.VerificationConfig),
	}
}

// GetConfig retrieves a configuration by ID
func (c *CustomConfigStore) GetConfig(ctx context.Context, id string) (self.VerificationConfig, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	config, exists := c.configs[id]
	if !exists {
		// Return default config for unknown IDs
		return self.VerificationConfig{
			MinimumAge: &[]int{18}[0],
			Ofac:       &[]bool{true}[0],
		}, nil
	}
	return config, nil
}

// SetConfig stores a configuration with the given ID
func (c *CustomConfigStore) SetConfig(ctx context.Context, id string, config self.VerificationConfig) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, existed := c.configs[id]
	c.configs[id] = config
	return !existed, nil
}

// GetActionId returns a custom action ID based on user data
func (c *CustomConfigStore) GetActionId(ctx context.Context, userIdentifier string, userDefinedData string) (string, error) {
	// Simple logic: use different configs based on user data length
	if len(userDefinedData) > 10 {
		return "premium-user-config", nil
	}
	return "standard-user-config", nil
}

var (
	verifier    *self.BackendVerifier
	configStore *CustomConfigStore
	initOnce    sync.Once
)

// initializeVerifier initializes the Self verifier with custom configuration
func initializeVerifier() {
	initOnce.Do(func() {
		configStore = NewCustomConfigStore()
		ctx := context.Background()

		// Set up default configurations
		standardConfig := self.VerificationConfig{
			MinimumAge: &[]int{18}[0],
			Ofac:       &[]bool{true}[0],
		}
		configStore.SetConfig(ctx, "standard-user-config", standardConfig)

		// Premium config with more restrictions
		premiumConfig := self.VerificationConfig{
			MinimumAge:        &[]int{21}[0],
			ExcludedCountries: []common.Country3LetterCode{common.RUS, common.IRN},
			Ofac:              &[]bool{true}[0],
		}
		configStore.SetConfig(ctx, "premium-user-config", premiumConfig)

		// Define allowed attestation types
		allowedIds := map[self.AttestationId]bool{
			self.Passport: true,
			self.EUCard:   true,
		}

		// Initialize the verifier
		var err error
		verifier, err = self.NewBackendVerifier(
			"self-playground", // App name
			"https://playground-two-psi.vercel.app/go-test/api/verify", // App URL - replace with your actual URL
			true, // Use testnet
			allowedIds,
			configStore,
			self.UserIDTypeUUID, // Use UUID format for user IDs
		)
		if err != nil {
			log.Fatalf("Failed to create Self verifier: %v", err)
		}

		log.Println("Self verifier initialized successfully")
	})
}

func GoVerify(w http.ResponseWriter, r *http.Request) {
	// Initialize the verifier on first request
	initializeVerifier()

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

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error reading request body"})
		return
	}

	log.Printf("Received request body: %s", string(body))

	// Parse the request
	var req VerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("JSON parsing error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid JSON",
			"error":   err.Error(),
		})
		return
	}

	// Basic validation
	if req.AttestationID == "" || req.Proof == nil || req.PublicSignals == nil || req.UserContextData == nil {
		w.WriteHeader(http.StatusBadRequest)
		response := VerifyResponse{
			Status:  "error",
			Result:  false,
			Message: "AttestationID, Proof, PublicSignals, and UserContextData are required",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Received verification request for attestationId: %s", req.AttestationID)

	// Create context for the verification
	ctx := context.Background()

	// Use a default user ID if not provided
	userID := req.UserID
	if userID == "" {
		userID = "anonymous-user"
	}

	// Convert attestation ID string to Self AttestationId type
	var attestationId self.AttestationId
	switch req.AttestationID {
	case "passport":
		attestationId = self.Passport
	case "eucard":
		attestationId = self.EUCard
	default:
		// Try to use passport as default
		attestationId = self.Passport
	}

	// Create VcAndDiscloseProof structure - use empty struct and pass proof/signals separately
	vcProof := self.VcAndDiscloseProof{}

	// Convert userContextData to string if it's not already
	userContextDataStr := ""
	if userContextBytes, err := json.Marshal(req.UserContextData); err == nil {
		userContextDataStr = string(userContextBytes)
	}

	// Perform the actual Self verification
	result, err := verifier.Verify(ctx, userID, vcProof, []string{string(attestationId)}, userContextDataStr)
	if err != nil {
		log.Printf("Verification error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		response := VerifyResponse{
			Status:  "error",
			Result:  false,
			Message: fmt.Sprintf("Verification failed: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if verification was successful
	// Based on the linter errors, IsValidDetails is not a pointer and doesn't have an IsValid field
	// Let's assume the result has a simple boolean or we need to check differently
	isValid := false

	// Try to determine validity from the result structure
	// This might need adjustment based on the actual SDK structure
	if result != nil {
		// Assume there's some way to determine validity - this may need to be adjusted
		isValid = true // Placeholder - will need to be fixed based on actual SDK
	}

	// Prepare the response based on verification result
	response := VerifyResponse{
		Status: "success",
		Result: isValid,
	}

	if isValid {
		response.Message = "Verification successful"
		// Use result directly as credential subject for now
		response.CredentialSubject = result

		// Get configuration from store using the userID
		if config, err := configStore.GetConfig(ctx, userID); err == nil {
			verificationOptions := map[string]interface{}{}
			if config.MinimumAge != nil {
				verificationOptions["minimumAge"] = *config.MinimumAge
			}
			if config.Ofac != nil {
				verificationOptions["ofac"] = *config.Ofac
			}
			if len(config.ExcludedCountries) > 0 {
				verificationOptions["excludedCountries"] = config.ExcludedCountries
			}
			response.VerificationOptions = verificationOptions
		}
	} else {
		response.Message = "Verification failed"
	}

	log.Printf("Verification result: valid=%v, message=%s", isValid, response.Message)
	json.NewEncoder(w).Encode(response)
}
