package sales

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type ShipEngineMock struct {
	ShipFunc func(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

func (m ShipEngineMock) Ship(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.ShipFunc != nil {
		return m.ShipFunc(ctx, moveID, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

type InvoiceEngineMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m InvoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{ContactID: request.ContactID}, nil
}

type InvoiceLookupMock struct {
	ListOpenByContactFunc func(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error)
}

func (m InvoiceLookupMock) ListOpenByContact(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error) {
	if m.ListOpenByContactFunc != nil {
		return m.ListOpenByContactFunc(ctx, contactID)
	}
	return []*accounting.Invoice{{Base: model.Base{ID: 1}}}, nil
}

type PaymentEngineMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error)
}

func (m PaymentEngineMock) Create(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Payment{ContactID: request.ContactID, Amount: request.Amount}, nil
}
