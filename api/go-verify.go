package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"playground/config"

	self "github.com/selfxyz/self/sdk/sdk-go"
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

// Handler is the equivalent of the TypeScript handler function (lines 37-55)
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var req VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate required fields - equivalent to TypeScript validation
		if req.Proof == nil || req.PublicSignals == nil || req.AttestationID == "" || req.UserContextData == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Proof, publicSignals, attestationId and userContextData are required",
			})
			return
		}

		// Convert req.Proof to self.VcAndDiscloseProof
		proofBytes, err := json.Marshal(req.Proof)
		if err != nil {
			log.Printf("Failed to marshal proof: %v", err)
			http.Error(w, "Invalid proof format", http.StatusBadRequest)
			return
		}

		var vcProof self.VcAndDiscloseProof
		if err := json.Unmarshal(proofBytes, &vcProof); err != nil {
			log.Printf("Failed to unmarshal proof to VcAndDiscloseProof: %v", err)
			http.Error(w, "Invalid proof structure", http.StatusBadRequest)
			return
		}

		// Convert req.PublicSignals to []string
		publicSignalsBytes, err := json.Marshal(req.PublicSignals)
		if err != nil {
			log.Printf("Failed to marshal public signals: %v", err)
			http.Error(w, "Invalid public signals format", http.StatusBadRequest)
			return
		}

		var publicSignals []string
		if err := json.Unmarshal(publicSignalsBytes, &publicSignals); err != nil {
			log.Printf("Failed to unmarshal public signals to []string: %v", err)
			http.Error(w, "Invalid public signals structure", http.StatusBadRequest)
			return
		}

		// Convert req.UserContextData to string
		userContextDataBytes, err := json.Marshal(req.UserContextData)
		if err != nil {
			log.Printf("Failed to marshal user context data: %v", err)
			http.Error(w, "Invalid user context data format", http.StatusBadRequest)
			return
		}
		userContextDataStr := string(userContextDataBytes)

		// Initialize config store - equivalent to TypeScript lines 52-55
		configStore, err := config.NewKVConfigStoreFromEnv()
		if err != nil {
			log.Printf("Failed to initialize config store: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Define allowed attestation types
		allowedIds := map[self.AttestationId]bool{
			self.Passport: true,
			self.EUCard:   true,
		}
		verifier, err := self.NewBackendVerifier(
			"self-playground-go",
			"https://playground-two-psi.vercel.app/api/go-verify",
			true, // Use testnet
			allowedIds,
			configStore,
			self.UserIDTypeUUID, // Use UUID format for user IDs
		)
		if err != nil {
			log.Printf("Failed to initialize verifier: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.Background()

		result, err := verifier.Verify(
			ctx,
			req.AttestationID,
			vcProof,
			publicSignals,
			userContextDataStr,
		)
		if err != nil {
			log.Printf("Verification failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(VerifyResponse{
				Status:  "error",
				Result:  false,
				Message: "Verification failed",
			})
			return
		}

		if result == nil || !result.IsValidDetails.IsValid {
			log.Printf("Verification failed - invalid result")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(VerifyResponse{
				Status:  "error",
				Result:  false,
				Message: "Verification failed",
			})
			return
		}

		// Get config from configStore - equivalent to TypeScript: configStore.getConfig(result.userData.userIdentifier)
		configResult, err := configStore.GetConfig(ctx, result.UserData.UserIdentifier)
		if err != nil {
			log.Printf("Failed to get config: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Type cast to SelfAppDisclosureConfig - equivalent to TypeScript: as unknown as SelfAppDisclosureConfig
		saveOptions := interface{}(configResult).(config.SelfAppDisclosureConfig)

		// Check if verification is valid - equivalent to TypeScript: if (result.isValidDetails.isValid)
		if result.IsValidDetails.IsValid {
			// Create filtered subject - equivalent to TypeScript: const filteredSubject = { ...result.discloseOutput };
			// Copy the struct to modify it
			filteredSubject := result.DiscloseOutput

			// Apply disclosure filters based on saveOptions - EXACT equivalent to TypeScript conditions

			// TypeScript: if (!saveOptions.issuing_state && filteredSubject)
			if saveOptions.IssuingState == nil || !*saveOptions.IssuingState {
				filteredSubject.IssuingState = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.name && filteredSubject)
			if saveOptions.Name == nil || !*saveOptions.Name {
				filteredSubject.Name = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.nationality && filteredSubject)
			if saveOptions.Nationality == nil || !*saveOptions.Nationality {
				filteredSubject.Nationality = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.date_of_birth && filteredSubject)
			if saveOptions.DateOfBirth == nil || !*saveOptions.DateOfBirth {
				filteredSubject.DateOfBirth = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.passport_number && filteredSubject)
			if saveOptions.PassportNumber == nil || !*saveOptions.PassportNumber {
				filteredSubject.IdNumber = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.gender && filteredSubject)
			if saveOptions.Gender == nil || !*saveOptions.Gender {
				filteredSubject.Gender = "Not disclosed"
			}

			// TypeScript: if (!saveOptions.expiry_date && filteredSubject)
			if saveOptions.ExpiryDate == nil || !*saveOptions.ExpiryDate {
				filteredSubject.ExpiryDate = "Not disclosed"
			}

			// Create excluded countries array with country code mapping (like TypeScript)
			var excludedCountriesForResponse []string
			if saveOptions.ExcludedCountries != nil {
				excludedCountriesForResponse = make([]string, len(saveOptions.ExcludedCountries))
				for i, countryCode := range saveOptions.ExcludedCountries {
					excludedCountriesForResponse[i] = string(countryCode)
				}
			}

			// Return successful verification result with filtered data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(VerifyResponse{
				Status:            "success",
				Result:            result.IsValidDetails.IsValid,
				CredentialSubject: filteredSubject,
				VerificationOptions: map[string]interface{}{
					"minimumAge":        saveOptions.MinimumAge,
					"ofac":              saveOptions.Ofac,
					"excludedCountries": excludedCountriesForResponse,
				},
			})
		} else {
			// Handle failed verification case - equivalent to TypeScript lines 127-134
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(VerifyResponse{
				Status:  "error",
				Result:  result.IsValidDetails.IsValid,
				Message: "Verification failed",
			})
		}
	}
}
