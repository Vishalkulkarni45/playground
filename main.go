package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	self "github.com/selfxyz/self/sdk/sdk-go"
	"github.com/selfxyz/self/sdk/sdk-go/common"
)

// CustomConfigStore implements a more sophisticated config store
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
	// In a real implementation, you might:
	// - Query a database
	// - Generate IDs based on user data
	// - Apply business logic

	// For this example, we'll create a simple mapping
	if len(userDefinedData) > 10 {
		return "premium-user-config", nil
	}
	return "standard-user-config", nil
}

// demonstrateVerification shows how to perform a verification (mock example)
func demonstrateVerification(verifier *self.BackendVerifier, configStore *CustomConfigStore) {
	fmt.Println("\n🔍 Demonstration of Verification Process:")

	ctx := context.Background()

	// Mock verification data (in real use, this would come from the client)
	mockProof := map[string]interface{}{
		"pi_a": []string{"123", "456"},
		"pi_b": [][]string{{"789", "012"}, {"345", "678"}},
		"pi_c": []string{"901", "234"},
	}

	mockPublicSignals := []interface{}{
		"1", "2", "3", "4", "5",
	}

	mockUserContextData := map[string]interface{}{
		"userIdentifier": "test-user-123",
		"timestamp":      "2024-01-01T00:00:00Z",
		"nonce":          "random-nonce-123",
	}

	// Convert to JSON strings as the API expects
	userContextDataStr := ""
	if userContextBytes, err := json.Marshal(mockUserContextData); err == nil {
		userContextDataStr = string(userContextBytes)
	}

	// Create VcAndDiscloseProof structure
	vcProof := self.VcAndDiscloseProof{}

	// Use variables to avoid unused variable warnings
	_ = mockProof
	_ = mockPublicSignals
	_ = vcProof

	// Mock verification call (this would normally verify real ZK proofs)
	fmt.Printf("📋 Mock verification for user: %s\n", "test-user-123")
	fmt.Printf("📄 Attestation type: %s\n", "passport")
	fmt.Printf("🔐 User context data: %s\n", userContextDataStr)

	// In a real scenario, you would call:
	// result, err := verifier.Verify(ctx, "test-user-123", vcProof, []string{"passport"}, userContextDataStr)

	// For demonstration, we'll show what the configuration lookup would do
	config, err := configStore.GetConfig(ctx, "test-user-123")
	if err == nil {
		fmt.Printf("⚙️  Applied configuration:\n")
		if config.MinimumAge != nil {
			fmt.Printf("   - Minimum age: %d\n", *config.MinimumAge)
		}
		if config.Ofac != nil {
			fmt.Printf("   - OFAC check: %v\n", *config.Ofac)
		}
		if len(config.ExcludedCountries) > 0 {
			fmt.Printf("   - Excluded countries: %v\n", config.ExcludedCountries)
		}
	}

	fmt.Println("✅ Verification process demonstrated!")
}

func main() {
	fmt.Println("🚀 Self SDK Custom Configuration Store Example")
	fmt.Println("===============================================")

	// Create custom config store
	configStore := NewCustomConfigStore()

	// Set up different configurations for different user types
	ctx := context.Background()

	// Standard user config (basic verification)
	standardConfig := self.VerificationConfig{
		MinimumAge: &[]int{18}[0],
		Ofac:       &[]bool{true}[0],
	}
	configStore.SetConfig(ctx, "standard-user-config", standardConfig)
	fmt.Println("📝 Standard user configuration created")

	// Premium user config (more restrictive)
	premiumConfig := self.VerificationConfig{
		MinimumAge:        &[]int{21}[0],
		ExcludedCountries: []common.Country3LetterCode{common.RUS, common.IRN},
		Ofac:              &[]bool{true}[0],
	}
	configStore.SetConfig(ctx, "premium-user-config", premiumConfig)
	fmt.Println("💎 Premium user configuration created")

	// Define allowed attestation types
	allowedIds := map[self.AttestationId]bool{
		self.Passport: true,
		self.EUCard:   true,
	}

	// Initialize the verifier
	verifier, err := self.NewBackendVerifier(
		"custom-config-app",          // App name
		"https://my-premium-app.com", // App URL
		true,                         // Use testnet
		allowedIds,
		configStore,
		self.UserIDTypeUUID, // Use UUID format for user IDs
	)
	if err != nil {
		log.Fatalf("❌ Failed to create verifier: %v", err)
	}

	fmt.Println("✅ Verifier with custom config store initialized!")

	// Demonstrate different configurations
	fmt.Println("\n📊 Configuration Examples:")

	// Test standard user config
	standardConfigResult, _ := configStore.GetConfig(ctx, "standard-user-config")
	fmt.Printf("👤 Standard users (min age: %d, OFAC: %v)\n",
		*standardConfigResult.MinimumAge, *standardConfigResult.Ofac)

	// Test premium user config
	premiumConfigResult, _ := configStore.GetConfig(ctx, "premium-user-config")
	fmt.Printf("💎 Premium users (min age: %d, excluded countries: %v, OFAC: %v)\n",
		*premiumConfigResult.MinimumAge, premiumConfigResult.ExcludedCountries, *premiumConfigResult.Ofac)

	// Demonstrate action ID generation
	fmt.Println("\n🎯 Action ID Examples:")

	shortData := "basic"
	actionId1, _ := configStore.GetActionId(ctx, "user123", shortData)
	fmt.Printf("📝 Short user data '%s' → Action ID: %s\n", shortData, actionId1)

	longData := "premium-user-with-extended-data"
	actionId2, _ := configStore.GetActionId(ctx, "user456", longData)
	fmt.Printf("📝 Long user data '%s' → Action ID: %s\n", longData, actionId2)

	// Demonstrate verification process
	demonstrateVerification(verifier, configStore)

	fmt.Println("\n🎉 Ready for verification with custom configuration logic!")
	fmt.Println("\n💡 To use this in production:")
	fmt.Println("   1. Replace mock data with real ZK proofs from clients")
	fmt.Println("   2. Implement proper error handling")
	fmt.Println("   3. Add logging and monitoring")
	fmt.Println("   4. Configure proper storage backend")
	fmt.Println("   5. Set up proper authentication and authorization")
}
