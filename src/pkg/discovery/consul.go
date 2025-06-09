package discovery

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceDiscovery interface for service registration and discovery
type ServiceDiscovery interface {
	RegisterService(serviceName, address, healthCheckPath string) error
	DeregisterService(serviceID string) error
	DiscoverService(serviceName string) ([]*ServiceEndpoint, error)
	DiscoverHealthyService(serviceName string) (*ServiceEndpoint, error)
}

// ServiceEndpoint represents a discovered service endpoint
type ServiceEndpoint struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

// ConsulClient implements ServiceDiscovery using Consul
type ConsulClient struct {
	client   *api.Client
	config   *ConsulConfig
	services map[string]string // Track registered services for cleanup
}

// ConsulConfig holds Consul configuration
type ConsulConfig struct {
	Address    string
	Datacenter string
	Token      string
	TTL        time.Duration
}

// NewConsulClient creates a new Consul service discovery client
func NewConsulClient(config *ConsulConfig) (*ConsulClient, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address
	consulConfig.Datacenter = config.Datacenter
	consulConfig.Token = config.Token

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulClient{
		client:   client,
		config:   config,
		services: make(map[string]string),
	}, nil
}

// RegisterService registers a service with Consul
func (c *ConsulClient) RegisterService(serviceName, address, healthCheckPath string) error {
	// Parse address to get host and port
	host, portStr, err := parseAddress(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port in address: %w", err)
	}

	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, host, port)

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    port,
		Address: host,
		Tags:    []string{"microservice", "autonomous"},
		Meta: map[string]string{
			"version":   "1.0.0",
			"framework": "go",
		},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s%s", address, healthCheckPath),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err = c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// Track the service for cleanup
	c.services[serviceName] = serviceID

	log.Printf("Service %s registered with Consul (ID: %s)", serviceName, serviceID)
	return nil
}

// DeregisterService removes a service from Consul
func (c *ConsulClient) DeregisterService(serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// Remove from tracking
	for name, id := range c.services {
		if id == serviceID {
			delete(c.services, name)
			break
		}
	}

	log.Printf("Service %s deregistered from Consul", serviceID)
	return nil
}

// DiscoverService finds all instances of a service
func (c *ConsulClient) DiscoverService(serviceName string) ([]*ServiceEndpoint, error) {
	services, _, err := c.client.Health().Service(serviceName, "", false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}

	endpoints := make([]*ServiceEndpoint, 0, len(services))
	for _, service := range services {
		endpoint := &ServiceEndpoint{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// DiscoverHealthyService finds a healthy instance of a service
func (c *ConsulClient) DiscoverHealthyService(serviceName string) (*ServiceEndpoint, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover healthy service: %w", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	// Return the first healthy instance
	service := services[0]
	return &ServiceEndpoint{
		ID:      service.Service.ID,
		Name:    service.Service.Service,
		Address: service.Service.Address,
		Port:    service.Service.Port,
		Tags:    service.Service.Tags,
		Meta:    service.Service.Meta,
	}, nil
}

// DeregisterAll deregisters all tracked services
func (c *ConsulClient) DeregisterAll() error {
	for _, serviceID := range c.services {
		if err := c.DeregisterService(serviceID); err != nil {
			log.Printf("Error deregistering service %s: %v", serviceID, err)
		}
	}
	return nil
}

// GetServiceURL returns the full URL for a service endpoint
func (e *ServiceEndpoint) GetServiceURL() string {
	return fmt.Sprintf("http://%s:%d", e.Address, e.Port)
}

// parseAddress splits an address into host and port
func parseAddress(address string) (host, port string, err error) {
	// Handle different address formats
	// Format: "host:port" or ":port" (localhost implied)
	if address[0] == ':' {
		return "localhost", address[1:], nil
	}

	// Find the last colon to split host and port
	lastColon := -1
	for i := len(address) - 1; i >= 0; i-- {
		if address[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon == -1 {
		return "", "", fmt.Errorf("invalid address format: %s", address)
	}

	return address[:lastColon], address[lastColon+1:], nil
}