package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mongocollectibles/rental-system/models"
)

// Config holds application configuration
type Config struct {
	PayMongoSecretKey string
	PayMongoPublicKey string
	ServerPort        string
	Environment       string
	Stores            []models.Store
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		PayMongoSecretKey: getEnv("TEST_SECRET_KEY", ""),
		PayMongoPublicKey: getEnv("TEST_PUBLIC_KEY", ""),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		Stores:            initializeStores(),
	}

	return config
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// initializeStores creates the default store locations
func initializeStores() []models.Store {
	return []models.Store{
		{
			ID:        "store-a",
			Name:      "MongoCollectibles Store A",
			Address:   "123 Main Street, Manila",
			Latitude:  14.5995,
			Longitude: 120.9842,
		},
		{
			ID:        "store-b",
			Name:      "MongoCollectibles Store B",
			Address:   "456 Quezon Avenue, Quezon City",
			Latitude:  14.6760,
			Longitude: 121.0437,
		},
		{
			ID:        "store-c",
			Name:      "MongoCollectibles Store C",
			Address:   "789 Makati Boulevard, Makati",
			Latitude:  14.5547,
			Longitude: 121.0244,
		},
	}
}

// GetStoreIndex returns the index of a store by ID
func (c *Config) GetStoreIndex(storeID string) int {
	for i, store := range c.Stores {
		if store.ID == storeID {
			return i
		}
	}
	return -1
}

// GetStoreByID returns a store by ID
func (c *Config) GetStoreByID(storeID string) *models.Store {
	for i := range c.Stores {
		if c.Stores[i].ID == storeID {
			return &c.Stores[i]
		}
	}
	return nil
}
