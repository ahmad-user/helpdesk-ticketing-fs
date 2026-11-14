package model

import "time"

// HealthStatus represents the health check status of the application and connected services.
type HealthStatus struct {
	Status    string            `json:"status"`
	App       string            `json:"app"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Database  string            `json:"database"`
	Details   map[string]string `json:"details,omitempty"`
}
