package services

import (
	"context"
	"fmt"
	"myproject/backend/db"
)

// ClientService handles client-related business logic
type ClientService struct {
	repo *db.Repository
}

// NewClientService creates a new client service
func NewClientService(repo *db.Repository) *ClientService {
	return &ClientService{repo: repo}
}

// Create creates a new client
func (s *ClientService) Create(ctx context.Context, client db.Client) (*db.Client, error) {
	// Validate required fields
	if client.Name == "" {
		return nil, fmt.Errorf("اسم العميل مطلوب") // Client name is required
	}

	// Validate email format if provided
	if client.Email != nil && *client.Email != "" {
		if !isValidEmail(*client.Email) {
			return nil, fmt.Errorf("تنسيق البريد الإلكتروني غير صحيح") // Invalid email format
		}
	}

	return s.repo.CreateClient(ctx, client)
}

// List retrieves clients with pagination and search
func (s *ClientService) List(ctx context.Context, query string, limit, offset int) ([]db.Client, int, error) {
	if limit <= 0 {
		limit = 20 // Default page size
	}
	if limit > 100 {
		limit = 100 // Max page size
	}

	return s.repo.ListClients(ctx, query, limit, offset)
}

// Get retrieves a client by ID
func (s *ClientService) Get(ctx context.Context, id int64) (*db.Client, error) {
	if id <= 0 {
		return nil, fmt.Errorf("معرف العميل غير صحيح") // Invalid client ID
	}

	client, err := s.repo.GetClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("العميل غير موجود") // Client not found
	}

	return client, nil
}

// Update updates an existing client
func (s *ClientService) Update(ctx context.Context, client db.Client) (*db.Client, error) {
	// Validate required fields
	if client.ID <= 0 {
		return nil, fmt.Errorf("معرف العميل مطلوب") // Client ID is required
	}
	if client.Name == "" {
		return nil, fmt.Errorf("اسم العميل مطلوب") // Client name is required
	}

	// Validate email format if provided
	if client.Email != nil && *client.Email != "" {
		if !isValidEmail(*client.Email) {
			return nil, fmt.Errorf("تنسيق البريد الإلكتروني غير صحيح") // Invalid email format
		}
	}

	// Check if client exists
	_, err := s.repo.GetClient(ctx, client.ID)
	if err != nil {
		return nil, fmt.Errorf("العميل غير موجود") // Client not found
	}

	return s.repo.UpdateClient(ctx, client)
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper email validation library
	return len(email) > 3 && len(email) < 255 &&
		contains(email, "@") && contains(email, ".")
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
