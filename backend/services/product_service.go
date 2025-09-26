package services

import (
	"context"
	"fmt"
	"barakaERP/backend/db"
)

// ProductService handles product-related business logic
type ProductService struct {
	repo *db.Repository
}

// NewProductService creates a new product service
func NewProductService(repo *db.Repository) *ProductService {
	return &ProductService{repo: repo}
}

// Create creates a new product
func (s *ProductService) Create(ctx context.Context, product db.Product) (*db.Product, error) {
	// Validate required fields
	if product.Name == "" {
		return nil, fmt.Errorf("اسم المنتج مطلوب") // Product name is required
	}

	// Validate price
	if err := db.ValidatePrice(product.UnitPriceCents); err != nil {
		return nil, fmt.Errorf("السعر غير صحيح: %v", err) // Invalid price
	}

	// Set default currency if not provided
	if product.Currency == "" {
		product.Currency = "DZD"
	}

	// Set default active status
	if !product.Active {
		product.Active = true
	}

	return s.repo.CreateProduct(ctx, product)
}

// List retrieves products with pagination and search
func (s *ProductService) List(ctx context.Context, query string, active *bool, limit, offset int) ([]db.Product, int, error) {
	if limit <= 0 {
		limit = 20 // Default page size
	}
	if limit > 100 {
		limit = 100 // Max page size
	}

	return s.repo.ListProducts(ctx, query, active, limit, offset)
}

// Get retrieves a product by ID
func (s *ProductService) Get(ctx context.Context, id int64) (*db.Product, error) {
	if id <= 0 {
		return nil, fmt.Errorf("معرف المنتج غير صحيح") // Invalid product ID
	}

	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("المنتج غير موجود") // Product not found
	}

	return product, nil
}

// Update updates an existing product
func (s *ProductService) Update(ctx context.Context, product db.Product) (*db.Product, error) {
	// Validate required fields
	if product.ID <= 0 {
		return nil, fmt.Errorf("معرف المنتج مطلوب") // Product ID is required
	}
	if product.Name == "" {
		return nil, fmt.Errorf("اسم المنتج مطلوب") // Product name is required
	}

	// Validate price
	if err := db.ValidatePrice(product.UnitPriceCents); err != nil {
		return nil, fmt.Errorf("السعر غير صحيح: %v", err) // Invalid price
	}

	// Set default currency if not provided
	if product.Currency == "" {
		product.Currency = "DZD"
	}

	// Check if product exists
	_, err := s.repo.GetProduct(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("المنتج غير موجود") // Product not found
	}

	return s.repo.UpdateProduct(ctx, product)
}

// Delete removes a product
func (s *ProductService) Delete(ctx context.Context, id int64) error {
	if id <= 0 { return fmt.Errorf("معرف المنتج غير صحيح") }
	if _, err := s.repo.GetProduct(ctx, id); err != nil { return fmt.Errorf("المنتج غير موجود") }
	// Check usage
	total, active, err := s.repo.ProductOrderUsageStats(ctx, id)
	if err != nil { return fmt.Errorf("تعذر التحقق من استخدام المنتج: %v", err) }
	if active > 0 {
		return fmt.Errorf("لا يمكن حذف المنتج لوجود طلبات غير ملغاة تستخدمه")
	}
	// If only canceled orders reference it, we can safely delete the product without touching historical canceled rows.
	// (Canceled orders/items will retain name/sku snapshots; product_id FK may prevent delete if still enforced.)
	// If FK constraints block, user must purge canceled orders first; we could add automatic purge later.
	if err := s.repo.DeleteProduct(ctx, id); err != nil { return err }
	_ = total // (Could be logged if needed)
	return nil
}

// ActivateProduct sets product as active
func (s *ProductService) ActivateProduct(ctx context.Context, id int64) (*db.Product, error) {
	product, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Active = true
	return s.repo.UpdateProduct(ctx, *product)
}

// DeactivateProduct sets product as inactive
func (s *ProductService) DeactivateProduct(ctx context.Context, id int64) (*db.Product, error) {
	product, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Active = false
	return s.repo.UpdateProduct(ctx, *product)
}
