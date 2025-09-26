package services

import (
	"context"
	"fmt"
	"barakaERP/backend/db"
	"strings"
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

	// Check if client exists
	_, err := s.repo.GetClient(ctx, client.ID)
	if err != nil {
		return nil, fmt.Errorf("العميل غير موجود") // Client not found
	}

	return s.repo.UpdateClient(ctx, client)
}

// AdjustDebt adjusts a client's debt by deltaCents (can be negative) and creates a debt payment record
func (s *ClientService) AdjustDebt(ctx context.Context, clientID int64, deltaCents int64, notes *string) (*db.Client, *db.DebtPayment, error) {
	if clientID <= 0 {
		return nil, nil, fmt.Errorf("معرف العميل غير صحيح")
	}
	
	// Check if client exists
	_, err := s.repo.GetClient(ctx, clientID)
	if err != nil {
		return nil, nil, fmt.Errorf("العميل غير موجود")
	}
	
	// Don't allow negative debt
	client, err := s.repo.GetClient(ctx, clientID)
	if err != nil {
		return nil, nil, fmt.Errorf("فشل في الحصول على معلومات العميل")
	}
	
	newDebt := client.DebtCents + deltaCents
	if newDebt < 0 {
		deltaCents = -client.DebtCents // Adjust delta to bring debt to 0
	}
	
	return s.repo.AdjustClientDebt(ctx, clientID, deltaCents, notes)
}

// GetDebtPayments retrieves all debt payment records with pagination
func (s *ClientService) GetDebtPayments(ctx context.Context, limit, offset int) (*db.PaginatedResult[db.DebtPaymentDetail], error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	return s.repo.GetDebtPayments(ctx, limit, offset)
}

// GetClientDebtPayments retrieves debt payment records for a specific client
func (s *ClientService) GetClientDebtPayments(ctx context.Context, clientID int64, limit, offset int) (*db.PaginatedResult[db.DebtPayment], error) {
	if clientID <= 0 {
		return nil, fmt.Errorf("معرف العميل غير صحيح")
	}
	
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	return s.repo.GetClientDebtPayments(ctx, clientID, limit, offset)
}

// Delete removes a client if no restricting relations block it
func (s *ClientService) Delete(ctx context.Context, id int64) error {
	if id <= 0 { return fmt.Errorf("معرف العميل غير صحيح") }
	// Check existence
	if _, err := s.repo.GetClient(ctx, id); err != nil { return fmt.Errorf("العميل غير موجود") }
	// First attempt delete
	// Ensure no active (non-canceled) orders remain
	if hasActive, errAct := s.repo.HasActiveOrdersForClient(ctx, id); errAct != nil {
		return fmt.Errorf("تعذر التحقق من الطلبات النشطة: %v", errAct)
	} else if hasActive {
		return fmt.Errorf("لا يمكن حذف العميل لوجود طلبات غير ملغاة")
	}
	if err := s.repo.DeleteClient(ctx, id); err != nil {
		// If FK constraint blocks, attempt to remove canceled orders then retry
		errStr := err.Error()
		if strings.Contains(errStr, "FOREIGN KEY") || strings.Contains(errStr, "foreign key") {
			// Remove canceled orders
			_, delErr := s.repo.DeleteCanceledOrdersForClient(ctx, id)
			if delErr != nil { return fmt.Errorf("تعذر حذف الطلبات الملغاة للعميل: %v", delErr) }
			// Retry delete
			if err2 := s.repo.DeleteClient(ctx, id); err2 != nil {
				return fmt.Errorf("لا يمكن حذف العميل لوجود طلبات أو فواتير نشطة")
			}
			return nil
		}
		return err
	}
	return nil
}
