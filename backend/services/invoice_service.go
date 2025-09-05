package services

import (
	"context"
	"fmt"
	"myproject/backend/db"
)

// InvoiceService handles invoice business logic
type InvoiceService struct {
	repo *db.Repository
}

func NewInvoiceService(repo *db.Repository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

func (s *InvoiceService) Create(ctx context.Context, draft db.InvoiceDraft) (*db.Invoice, error) {
	if draft.ClientID <= 0 {
		return nil, fmt.Errorf("client id required")
	}
	if len(draft.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}
	for i, it := range draft.Items {
		if it.Qty <= 0 {
			return nil, fmt.Errorf("quantity must be > 0 for item %d", i+1)
		}
		if it.UnitPriceCents <= 0 {
			return nil, fmt.Errorf("unit price must be > 0 for item %d", i+1)
		}
	}
	return s.repo.CreateInvoice(ctx, draft)
}

func (s *InvoiceService) List(ctx context.Context, limit, offset int) ([]db.InvoiceDetail, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListInvoices(ctx, limit, offset)
}

func (s *InvoiceService) Get(ctx context.Context, id int64) (*db.InvoiceDetail, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid invoice id")
	}
	return s.repo.GetInvoiceDetail(ctx, id)
}
