package services

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// LicenseService handles license validation
type LicenseService struct {
	licenseURL    string
	client        *http.Client
	checkEnabled  bool
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return // .env file doesn't exist, skip
	}

	file, err := os.Open(envFile)
	if err != nil {
		return // Can't open file, skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Invalid format, skip
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// NewLicenseService creates a new license service
func NewLicenseService() *LicenseService {
	// Load .env file
	loadEnvFile()
	
	// Check if license checking is enabled
	checkEnabled := true
	if env := os.Getenv("ENABLE_LICENSE_CHECK"); env != "" {
		checkEnabled = strings.ToLower(env) == "true"
	}

	return &LicenseService{
		licenseURL:   "https://api-sadine.sorino.io/test",
		checkEnabled: checkEnabled,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LicenseStatus represents the license validation result
type LicenseStatus struct {
	IsValid   bool   `json:"is_valid"`
	Message   string `json:"message"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ValidateLicense checks if the license is valid by making a GET request to the license server
func (s *LicenseService) ValidateLicense() (*LicenseStatus, error) {
	// If license checking is disabled, always return valid
	if !s.checkEnabled {
		return &LicenseStatus{
			IsValid: true,
			Message: "License checking is disabled in development mode",
		}, nil
	}

	resp, err := s.client.Get(s.licenseURL)
	if err != nil {
		// Network error - assume trial expired for security
		return &LicenseStatus{
			IsValid: false,
			Message: "Unable to validate license. Please check your internet connection.",
		}, nil
	}
	defer resp.Body.Close()

	// If we get a 404, the trial has ended
	if resp.StatusCode == http.StatusNotFound {
		return &LicenseStatus{
			IsValid: false,
			Message: "Your free trial has ended. Please contact support to continue using barakaERP.",
		}, nil
	}

	// If we get a 200, the trial is still active
	if resp.StatusCode == http.StatusOK {
		return &LicenseStatus{
			IsValid: true,
			Message: "Trial license is active",
		}, nil
	}

	// Any other status code - assume trial expired for security
	return &LicenseStatus{
		IsValid: false,
		Message: fmt.Sprintf("License validation failed with status: %d", resp.StatusCode),
	}, nil
}

// CheckLicenseQuiet performs a license check without detailed error messages
func (s *LicenseService) CheckLicenseQuiet() bool {
	// If license checking is disabled, always return true
	if !s.checkEnabled {
		return true
	}

	status, err := s.ValidateLicense()
	if err != nil {
		return false
	}
	return status.IsValid
}