package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConfigManager provides multi-layer configuration management
type ConfigManager struct {
	consulClient *api.Client
	cache        map[string]string
	cacheMutex   sync.RWMutex
	prefix       string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(consulAddr, prefix string) (*ConfigManager, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConfigManager{
		consulClient: client,
		cache:        make(map[string]string),
		prefix:       prefix,
	}, nil
}

// GetString gets a string configuration value with fallback hierarchy:
// 1. Environment variable
// 2. Consul KV store
// 3. Default value
func (c *ConfigManager) GetString(key, defaultValue string) string {
	// Try environment variable first
	if envValue := os.Getenv(key); envValue != "" {
		return envValue
	}

	// Try Consul KV store
	if consulValue := c.getFromConsul(key); consulValue != "" {
		return consulValue
	}

	return defaultValue
}

// GetInt gets an integer configuration value
func (c *ConfigManager) GetInt(key string, defaultValue int) int {
	stringValue := c.GetString(key, strconv.Itoa(defaultValue))
	if intValue, err := strconv.Atoi(stringValue); err == nil {
		return intValue
	}
	return defaultValue
}

// GetBool gets a boolean configuration value
func (c *ConfigManager) GetBool(key string, defaultValue bool) bool {
	stringValue := c.GetString(key, strconv.FormatBool(defaultValue))
	if boolValue, err := strconv.ParseBool(stringValue); err == nil {
		return boolValue
	}
	return defaultValue
}

// GetDuration gets a duration configuration value
func (c *ConfigManager) GetDuration(key string, defaultValue time.Duration) time.Duration {
	stringValue := c.GetString(key, defaultValue.String())
	if duration, err := time.ParseDuration(stringValue); err == nil {
		return duration
	}
	return defaultValue
}

// GetStringSlice gets a string slice configuration value (comma-separated)
func (c *ConfigManager) GetStringSlice(key string, defaultValue []string) []string {
	stringValue := c.GetString(key, strings.Join(defaultValue, ","))
	if stringValue == "" {
		return defaultValue
	}
	return strings.Split(stringValue, ",")
}

// SetInConsul sets a configuration value in Consul KV store
func (c *ConfigManager) SetInConsul(key, value string) error {
	fullKey := c.getFullKey(key)
	
	pair := &api.KVPair{
		Key:   fullKey,
		Value: []byte(value),
	}

	_, err := c.consulClient.KV().Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to set config in consul: %w", err)
	}

	// Update cache
	c.cacheMutex.Lock()
	c.cache[key] = value
	c.cacheMutex.Unlock()

	return nil
}

// DeleteFromConsul removes a configuration value from Consul KV store
func (c *ConfigManager) DeleteFromConsul(key string) error {
	fullKey := c.getFullKey(key)
	
	_, err := c.consulClient.KV().Delete(fullKey, nil)
	if err != nil {
		return fmt.Errorf("failed to delete config from consul: %w", err)
	}

	// Remove from cache
	c.cacheMutex.Lock()
	delete(c.cache, key)
	c.cacheMutex.Unlock()

	return nil
}

// RefreshCache refreshes the configuration cache from Consul
func (c *ConfigManager) RefreshCache() error {
	pairs, _, err := c.consulClient.KV().List(c.prefix, nil)
	if err != nil {
		return fmt.Errorf("failed to refresh config cache: %w", err)
	}

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// Clear cache
	c.cache = make(map[string]string)

	// Populate cache
	for _, pair := range pairs {
		key := strings.TrimPrefix(pair.Key, c.prefix+"/")
		c.cache[key] = string(pair.Value)
	}

	return nil
}

// WatchKey watches for changes to a specific configuration key
func (c *ConfigManager) WatchKey(key string, callback func(string, string)) error {
	fullKey := c.getFullKey(key)
	
	go func() {
		var lastIndex uint64
		
		for {
			pair, meta, err := c.consulClient.KV().Get(fullKey, &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  time.Minute * 5,
			})
			
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
			
			lastIndex = meta.LastIndex
			
			if pair != nil {
				newValue := string(pair.Value)
				
				// Update cache
				c.cacheMutex.Lock()
				oldValue := c.cache[key]
				c.cache[key] = newValue
				c.cacheMutex.Unlock()
				
				// Call callback if value changed
				if oldValue != newValue {
					callback(key, newValue)
				}
			}
		}
	}()
	
	return nil
}

// GetAll returns all configuration values from cache
func (c *ConfigManager) GetAll() map[string]string {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	
	result := make(map[string]string)
	for k, v := range c.cache {
		result[k] = v
	}
	
	return result
}

// getFromConsul retrieves a value from Consul KV store with caching
func (c *ConfigManager) getFromConsul(key string) string {
	// Check cache first
	c.cacheMutex.RLock()
	if value, exists := c.cache[key]; exists {
		c.cacheMutex.RUnlock()
		return value
	}
	c.cacheMutex.RUnlock()

	// Fetch from Consul
	fullKey := c.getFullKey(key)
	pair, _, err := c.consulClient.KV().Get(fullKey, nil)
	if err != nil || pair == nil {
		return ""
	}

	value := string(pair.Value)

	// Update cache
	c.cacheMutex.Lock()
	c.cache[key] = value
	c.cacheMutex.Unlock()

	return value
}

// getFullKey constructs the full Consul key with prefix
func (c *ConfigManager) getFullKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s/%s", c.prefix, key)
}

// ServiceConfig provides service-specific configuration
type ServiceConfig struct {
	ServiceName string
	Port        int
	ConsulAddr  string
	RedisAddr   string
	ConfigMgr   *ConfigManager
}

// NewServiceConfig creates a service configuration
func NewServiceConfig(serviceName string) (*ServiceConfig, error) {
	port := getEnvInt(fmt.Sprintf("%s_PORT", strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_"))), 8080)
	consulAddr := getEnvString("CONSUL_ADDR", "consul:8500")
	redisAddr := getEnvString("REDIS_ADDR", "redis:6379")

	configMgr, err := NewConfigManager(consulAddr, fmt.Sprintf("config/%s", serviceName))
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	return &ServiceConfig{
		ServiceName: serviceName,
		Port:        port,
		ConsulAddr:  consulAddr,
		RedisAddr:   redisAddr,
		ConfigMgr:   configMgr,
	}, nil
}

// Helper functions for environment variable handling
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}