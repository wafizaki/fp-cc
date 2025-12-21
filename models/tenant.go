package models

import (
	"fmt"
	"regexp"
	"slices"
	"time"
)

type Tenant struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Port          int       `json:"port"`
	ContainerName string    `json:"container_name"`
	VolumeName    string    `json:"volume_name"`
	Status        string    `json:"status"`
	URL           string    `json:"url"`
	Password      string    `json:"password,omitempty"` // Only populated when fetched from logs
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTenantRequest struct {
	Status string `json:"status"`
}

func (r *CreateTenantRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("tenant name is required")
	}

	matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", r.Name)
	if err != nil {
		return fmt.Errorf("failed to validate tenant name format: %w", err)
	}

	if !matched {
		return fmt.Errorf("tenant name must contain only alphanumeric characters and underscores")
	}

	if len(r.Name) < 3 {
		return fmt.Errorf("tenant name must be at least 3 characters long")
	}

	if len(r.Name) > 50 {
		return fmt.Errorf("tenant name must not exceed 50 characters")
	}

	return nil
}

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusError   = "error"
	StatusDeleted = "deleted"
)

func IsValidStatus(status string) bool {
	validStatuses := []string{StatusRunning, StatusStopped, StatusError, StatusDeleted}
	return slices.Contains(validStatuses, status)
}

