package services

import (
	"context"
	"fmt"
	"myproject/backend/db"
)

// OrderService handles order-related business logic
type OrderService struct {
	repo *db.Repository
}

// NewOrderService creates a new order service
func NewOrderService(repo *db.Repository) *OrderService {
	return &OrderService{repo: repo}
}

// Create creates a new order
func (s *OrderService) Create(ctx context.Context, draft db.OrderDraft) (*db.Order, error) {
	// Validate required fields
	if draft.ClientID <= 0 {
		return nil, fmt.Errorf("معرف العميل مطلوب") // Client ID is required
	}

	if len(draft.Items) == 0 {
		return nil, fmt.Errorf("يجب إضافة عنصر واحد على الأقل للطلب") // At least one item is required
	}

	// Validate items
	for i, item := range draft.Items {
		if item.Qty <= 0 {
			return nil, fmt.Errorf("الكمية يجب أن تكون أكبر من صفر للعنصر %d", i+1) // Quantity must be greater than zero
		}
		if item.UnitPriceCents <= 0 {
			return nil, fmt.Errorf("سعر الوحدة يجب أن يكون أكبر من صفر للعنصر %d", i+1) // Unit price must be greater than zero
		}
		if item.NameSnapshot == "" {
			return nil, fmt.Errorf("اسم المنتج مطلوب للعنصر %d", i+1) // Product name is required
		}
		if item.Currency == "" {
			item.Currency = "USD" // Default currency
		}
	}

	// Set default values
	if draft.DiscountPercent < 0 || draft.DiscountPercent > 100 {
		draft.DiscountPercent = 0
	}

	// Verify client exists
	_, err := s.repo.GetClient(ctx, draft.ClientID)
	if err != nil {
		return nil, fmt.Errorf("العميل غير موجود") // Client not found
	}

	return s.repo.CreateOrder(ctx, draft)
}

// List retrieves orders with pagination and filters
func (s *OrderService) List(ctx context.Context, filters db.OrderFilters, limit, offset int) ([]db.OrderDetail, int, error) {
	if limit <= 0 {
		limit = 20 // Default page size
	}
	if limit > 100 {
		limit = 100 // Max page size
	}

	return s.repo.ListOrders(ctx, filters, limit, offset)
}

// Get retrieves an order by ID with details
func (s *OrderService) Get(ctx context.Context, id int64) (*db.OrderDetail, error) {
	if id <= 0 {
		return nil, fmt.Errorf("معرف الطلب غير صحيح") // Invalid order ID
	}

	order, err := s.repo.GetOrderDetail(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("الطلب غير موجود") // Order not found
	}

	return order, nil
}

// Update updates an existing order
func (s *OrderService) Update(ctx context.Context, update db.OrderUpdate) (*db.Order, error) {
	if update.ID <= 0 {
		return nil, fmt.Errorf("معرف الطلب مطلوب") // Order ID is required
	}

	// Check if order exists
	_, err := s.repo.GetOrderDetail(ctx, update.ID)
	if err != nil {
		return nil, fmt.Errorf("الطلب غير موجود") // Order not found
	}

	// Validate items if provided
	if len(update.Items) > 0 {
		for i, item := range update.Items {
			if item.Qty <= 0 {
				return nil, fmt.Errorf("الكمية يجب أن تكون أكبر من صفر للعنصر %d", i+1) // Quantity must be greater than zero
			}
			if item.UnitPriceCents <= 0 {
				return nil, fmt.Errorf("سعر الوحدة يجب أن يكون أكبر من صفر للعنصر %d", i+1) // Unit price must be greater than zero
			}
			if item.NameSnapshot == "" {
				return nil, fmt.Errorf("اسم المنتج مطلوب للعنصر %d", i+1) // Product name is required
			}
			if item.Currency == "" {
				item.Currency = "USD" // Default currency
			}
		}
	}

	// Validate discount and tax percentages
	if update.DiscountPercent != nil && (*update.DiscountPercent < 0 || *update.DiscountPercent > 100) {
		return nil, fmt.Errorf("نسبة الخصم يجب أن تكون بين 0 و 100") // Discount percentage must be between 0 and 100
	}

	return s.repo.UpdateOrder(ctx, update)
}

// Delete deletes an order (soft delete by setting status to CANCELED)
func (s *OrderService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("معرف الطلب غير صحيح") // Invalid order ID
	}

	// Check if order exists
	orderDetail, err := s.repo.GetOrderDetail(ctx, id)
	if err != nil {
		return fmt.Errorf("الطلب غير موجود") // Order not found
	}

	// Don't allow deletion of completed orders
	if orderDetail.Order.Status == db.OrderStatusCompleted {
		return fmt.Errorf("لا يمكن حذف طلب مكتمل") // Cannot delete completed order
	}

	// Cancel the order instead of hard delete
	status := db.OrderStatusCanceled
	update := db.OrderUpdate{
		ID:     id,
		Status: &status,
	}

	_, err = s.repo.UpdateOrder(ctx, update)
	return err
}

// GetOrderStatuses returns available order statuses
func (s *OrderService) GetOrderStatuses() []string {
	return []string{
		db.OrderStatusPending,
		db.OrderStatusConfirmed,
		db.OrderStatusCompleted,
		db.OrderStatusCanceled,
	}
}
