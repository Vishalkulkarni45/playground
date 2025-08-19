package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	self "github.com/selfxyz/self/sdk/sdk-go"
	"github.com/selfxyz/self/sdk/sdk-go/common"
)

// SelfAppDisclosureConfig matches the TypeScript interface exactly
// This is the Go equivalent of the SelfAppDisclosureConfig interface
type SelfAppDisclosureConfig struct {
	IssuingState      *bool                       `json:"issuing_state,omitempty"`
	Name              *bool                       `json:"name,omitempty"`
	PassportNumber    *bool                       `json:"passport_number,omitempty"`
	Nationality       *bool                       `json:"nationality,omitempty"`
	DateOfBirth       *bool                       `json:"date_of_birth,omitempty"`
	Gender            *bool                       `json:"gender,omitempty"`
	ExpiryDate        *bool                       `json:"expiry_date,omitempty"`
	Ofac              *bool                       `json:"ofac,omitempty"`
	ExcludedCountries []common.Country3LetterCode `json:"excludedCountries,omitempty"`
	MinimumAge        *int                        `json:"minimumAge,omitempty"`
}

// KVConfigStore implements a Redis-based configuration store for Self verification
// This is the Go equivalent of the TypeScript KVConfigStore class
type KVConfigStore struct {
	redis *redis.Client
}

// NewKVConfigStore creates a new Redis-based config store
// Equivalent to the TypeScript constructor that takes url and token
func NewKVConfigStore(redisURL, redisToken string) (*KVConfigStore, error) {
	// Parse Redis connection from URL and token
	// For Upstash Redis, the URL format is typically: redis://default:token@host:port
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Set the password from token if provided
	if redisToken != "" {
		opt.Password = redisToken
	}

	client := redis.NewClient(opt)

	// Test the connection
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &KVConfigStore{
		redis: client,
	}, nil
}

// NewKVConfigStoreFromEnv creates a new Redis-based config store using environment variables
// This matches the TypeScript version that uses process.env variables
func NewKVConfigStoreFromEnv() (*KVConfigStore, error) {
	redisURL := os.Getenv("KV_REST_API_URL")
	redisToken := os.Getenv("KV_REST_API_TOKEN")

	if redisURL == "" {
		return nil, fmt.Errorf("KV_REST_API_URL environment variable is required")
	}
	if redisToken == "" {
		return nil, fmt.Errorf("KV_REST_API_TOKEN environment variable is required")
	}

	return NewKVConfigStore(redisURL, redisToken)
}

func (kv *KVConfigStore) GetActionId(ctx context.Context, userIdentifier string, userDefinedData string) (string, error) {
	return userIdentifier, nil
}

func (kv *KVConfigStore) SetConfig(ctx context.Context, id string, config self.VerificationConfig) (bool, error) {
	// Serialize the config to JSON, just like the TypeScript version: JSON.stringify(config)
	configJSON, err := json.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("failed to marshal config: %w", err)
	}

	err = kv.redis.Set(ctx, id, string(configJSON), 0).Err()
	if err != nil {
		return false, fmt.Errorf("failed to set config in Redis: %w", err)
	}

	return true, nil
}

// SetWithExpiration stores a key-value pair with expiration, matching TypeScript kv.set(key, value, { ex: seconds })
func (kv *KVConfigStore) SetWithExpiration(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := kv.redis.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key with expiration in Redis: %w", err)
	}
	return nil
}

func (kv *KVConfigStore) GetConfig(ctx context.Context, id string) (self.VerificationConfig, error) {
	// Get from Redis - this matches: await this.redis.get(id)
	configJSON, err := kv.redis.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist - return default config
			return self.VerificationConfig{
				MinimumAge: &[]int{18}[0],
				Ofac:       &[]bool{true}[0],
			}, nil
		}
		return self.VerificationConfig{}, fmt.Errorf("failed to get config from Redis: %w", err)
	}

	var config self.VerificationConfig
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		return self.VerificationConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// Close closes the Redis connection
func (kv *KVConfigStore) Close() error {
	return kv.redis.Close()
}
