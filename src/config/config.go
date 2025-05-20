package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port int
	Host string

	// Database configuration
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Authentication
	JWTSecret string
	JWTExpiry int

	// LLM configuration
	LLMAPIKey         string
	LLMModel          string
	LLMMaxTokens      int
	LLMTemperature    float64
	ContextWindowSize int

	// Search configuration
	SearchAPIKey string
	SearchURL    string

	// Plagiarism API configuration
	PlagiarismAPIKey  string
	PlagiarismAPIURL  string
	EnablePlagiarism  bool
	EnableFactChecking bool
	EnableSEO         bool
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Initialize config with defaults
	config := &Config{
		Port:              8080,
		Host:              "0.0.0.0",
		DBPort:            5432,
		DBSSLMode:         "disable",
		JWTExpiry:         24 * 60, // 24 hours in minutes
		LLMModel:          "gpt-4",
		LLMMaxTokens:      2048,
		LLMTemperature:    0.7,
		ContextWindowSize: 8192,
		EnablePlagiarism:  true,
		EnableFactChecking: true,
		EnableSEO:         true,
	}

	// Server config
	if port, err := strconv.Atoi(getEnv("PORT", "8080")); err == nil {
		config.Port = port
	}
	config.Host = getEnv("HOST", "0.0.0.0")

	// Database config
	config.DBHost = getEnv("DB_HOST", "localhost")
	if port, err := strconv.Atoi(getEnv("DB_PORT", "5432")); err == nil {
		config.DBPort = port
	}
	config.DBUser = getEnv("DB_USER", "postgres")
	config.DBPassword = getEnv("DB_PASSWORD", "")
	config.DBName = getEnv("DB_NAME", "contentservice")
	config.DBSSLMode = getEnv("DB_SSLMODE", "disable")

	// Authentication config
	config.JWTSecret = getEnv("JWT_SECRET", "")
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if expiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "1440")); err == nil {
		config.JWTExpiry = expiry
	}

	// LLM config
	config.LLMAPIKey = getEnv("LLM_API_KEY", "")
	if config.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY is required")
	}
	config.LLMModel = getEnv("LLM_MODEL", "gpt-4")
	if tokens, err := strconv.Atoi(getEnv("LLM_MAX_TOKENS", "2048")); err == nil {
		config.LLMMaxTokens = tokens
	}
	if temp, err := strconv.ParseFloat(getEnv("LLM_TEMPERATURE", "0.7"), 64); err == nil {
		config.LLMTemperature = temp
	}
	if windowSize, err := strconv.Atoi(getEnv("CONTEXT_WINDOW_SIZE", "8192")); err == nil {
		config.ContextWindowSize = windowSize
	}

	// Search config
	config.SearchAPIKey = getEnv("SEARCH_API_KEY", "")
	config.SearchURL = getEnv("SEARCH_URL", "")

	// Plagiarism API config
	config.PlagiarismAPIKey = getEnv("PLAGIARISM_API_KEY", "")
	config.PlagiarismAPIURL = getEnv("PLAGIARISM_API_URL", "")
	config.EnablePlagiarism = getBoolEnv("ENABLE_PLAGIARISM", true)
	config.EnableFactChecking = getBoolEnv("ENABLE_FACT_CHECKING", true)
	config.EnableSEO = getBoolEnv("ENABLE_SEO", true)

	return config, nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getBoolEnv gets a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value := strings.ToLower(valueStr)
	return value == "true" || value == "yes" || value == "1"
}
