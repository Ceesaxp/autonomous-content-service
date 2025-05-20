package entities

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// ClientStatus represents the status of a client in the system
type ClientStatus string

const (
	ClientStatusActive    ClientStatus = "Active"
	ClientStatusInactive  ClientStatus = "Inactive"
	ClientStatusPending   ClientStatus = "Pending"
	ClientStatusSuspended ClientStatus = "Suspended"
)

// Address represents a physical address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// Validate ensures the address has all required fields
func (a Address) Validate() error {
	if a.Street == "" {
		return errors.New("street is required")
	}
	if a.City == "" {
		return errors.New("city is required")
	}
	if a.Country == "" {
		return errors.New("country is required")
	}
	return nil
}

// Client represents a client of the content creation service
type Client struct {
	ClientID      uuid.UUID     `json:"clientId"`
	Name          string        `json:"name"`
	ContactEmail  string        `json:"contactEmail"`
	ContactPhone  string        `json:"contactPhone"`
	BillingAddress Address       `json:"billingAddress"`
	Timezone      string        `json:"timezone"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
	Status        ClientStatus  `json:"status"`
	Profile       *ClientProfile `json:"profile,omitempty"`
}

// NewClient creates a new client with the given properties
func NewClient(name, email, phone string, address Address, timezone string) (*Client, error) {
	c := &Client{
		ClientID:      uuid.New(),
		Name:          name,
		ContactEmail:  email,
		ContactPhone:  phone,
		BillingAddress: address,
		Timezone:      timezone,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Status:        ClientStatusPending,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate ensures the client has all required fields and valid data
func (c *Client) Validate() error {
	if c.Name == "" {
		return errors.New("client name is required")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(c.ContactEmail) {
		return errors.New("invalid email format")
	}

	// Validate phone number (simple validation)
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	if !phoneRegex.MatchString(c.ContactPhone) {
		return errors.New("invalid phone number format")
	}

	// Validate address
	if err := c.BillingAddress.Validate(); err != nil {
		return err
	}

	return nil
}

// Activate changes the client status to Active
func (c *Client) Activate() {
	c.Status = ClientStatusActive
	c.UpdateTimestamp()
}

// Deactivate changes the client status to Inactive
func (c *Client) Deactivate() {
	c.Status = ClientStatusInactive
	c.UpdateTimestamp()
}

// Suspend changes the client status to Suspended
func (c *Client) Suspend() {
	c.Status = ClientStatusSuspended
	c.UpdateTimestamp()
}

// UpdateTimestamp updates the UpdatedAt timestamp to the current time
func (c *Client) UpdateTimestamp() {
	c.UpdatedAt = time.Now()
}

// IsActive returns true if the client status is Active
func (c *Client) IsActive() bool {
	return c.Status == ClientStatusActive
}
