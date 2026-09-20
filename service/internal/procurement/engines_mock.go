package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type OfferEngineMock struct {
	BestOfferForSupplierFunc func(ctx context.Context, variantID, supplierID uint64, qty amount.Amount, date time.Time) (*products.SupplierProduct, error)
}

func (m OfferEngineMock) BestOfferForSupplier(ctx context.Context, variantID, supplierID uint64, qty amount.Amount, date time.Time) (*products.SupplierProduct, error) {
	if m.BestOfferForSupplierFunc != nil {
		return m.BestOfferForSupplierFunc(ctx, variantID, supplierID, qty, date)
	}
	return &products.SupplierProduct{ItemID: variantID, SupplierID: supplierID}, nil
}

type SupplierProfileLookupMock struct {
	FindByContactFunc func(ctx context.Context, contactID uint64) (*contacts.SupplierProfile, error)
}

func (m SupplierProfileLookupMock) FindByContact(ctx context.Context, contactID uint64) (*contacts.SupplierProfile, error) {
	if m.FindByContactFunc != nil {
		return m.FindByContactFunc(ctx, contactID)
	}
	return &contacts.SupplierProfile{ContactID: contactID, Active: true}, nil
}

type ReceiveEngineMock struct {
	ReceiveFunc func(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

func (m ReceiveEngineMock) Receive(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.ReceiveFunc != nil {
		return m.ReceiveFunc(ctx, moveID, unitCost, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

type ApprovalEngineMock struct {
	IsApprovedFunc func(ctx context.Context, ownerType string, ownerID uint64) (bool, error)
	CreateFunc     func(ctx context.Context, organizationID uint64, ownerType string, ownerID, requestedBy uint64, approverIDs []uint64) (*crosscutting.ApprovalRequest, error)
	StateFunc      func(ctx context.Context, ownerType string, ownerID uint64) (*crosscutting.ApprovalRequest, []*crosscutting.ApprovalStep, error)
}

func (m ApprovalEngineMock) IsApproved(ctx context.Context, ownerType string, ownerID uint64) (bool, error) {
	if m.IsApprovedFunc != nil {
		return m.IsApprovedFunc(ctx, ownerType, ownerID)
	}
	return false, nil
}

func (m ApprovalEngineMock) Create(ctx context.Context, organizationID uint64, ownerType string, ownerID, requestedBy uint64, approverIDs []uint64) (*crosscutting.ApprovalRequest, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, organizationID, ownerType, ownerID, requestedBy, approverIDs)
	}
	return &crosscutting.ApprovalRequest{OrganizationID: organizationID, OwnerType: ownerType, OwnerID: ownerID, RequestedBy: requestedBy}, nil
}

func (m ApprovalEngineMock) State(ctx context.Context, ownerType string, ownerID uint64) (*crosscutting.ApprovalRequest, []*crosscutting.ApprovalStep, error) {
	if m.StateFunc != nil {
		return m.StateFunc(ctx, ownerType, ownerID)
	}
	return nil, []*crosscutting.ApprovalStep{}, nil
}

type BillEngineMock struct {
	CreateSupplierBillFunc func(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error)
	CreateCreditNoteFunc   func(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}

func (m BillEngineMock) CreateSupplierBill(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
	if m.CreateSupplierBillFunc != nil {
		return m.CreateSupplierBillFunc(ctx, request)
	}
	return &accounting.Invoice{ContactID: request.ContactID}, nil
}

func (m BillEngineMock) CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	if m.CreateCreditNoteFunc != nil {
		return m.CreateCreditNoteFunc(ctx, request)
	}
	return &accounting.Invoice{ContactID: 0}, nil
}

type OpenInvoiceLookupMock struct {
	ListOpenByContactFunc func(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error)
}

func (m OpenInvoiceLookupMock) ListOpenByContact(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error) {
	if m.ListOpenByContactFunc != nil {
		return m.ListOpenByContactFunc(ctx, contactID)
	}
	return []*accounting.Invoice{}, nil
}

type OutboundPaymentEngineMock struct {
	CreateOutboundFunc func(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error)
}

func (m OutboundPaymentEngineMock) CreateOutbound(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error) {
	if m.CreateOutboundFunc != nil {
		return m.CreateOutboundFunc(ctx, request)
	}
	return &accounting.Payment{ContactID: request.ContactID, Amount: request.Amount}, nil
}

type QualityEngineMock struct {
	TriggerChecksFunc   func(ctx context.Context, organizationID uint64, shipmentID uint64, itemIDs []uint64) (int, error)
	HasFailedChecksFunc func(ctx context.Context, shipmentID uint64) (bool, error)
}

func (m QualityEngineMock) TriggerChecks(ctx context.Context, organizationID uint64, shipmentID uint64, itemIDs []uint64) (int, error) {
	if m.TriggerChecksFunc != nil {
		return m.TriggerChecksFunc(ctx, organizationID, shipmentID, itemIDs)
	}
	return 0, nil
}

func (m QualityEngineMock) HasFailedChecks(ctx context.Context, shipmentID uint64) (bool, error) {
	if m.HasFailedChecksFunc != nil {
		return m.HasFailedChecksFunc(ctx, shipmentID)
	}
	return false, nil
}

type ExpenseAccountEngineMock struct {
	ResolveExpenseAccountFunc func(ctx context.Context, variantID uint64) (uint64, error)
}

func (m ExpenseAccountEngineMock) ResolveExpenseAccount(ctx context.Context, variantID uint64) (uint64, error) {
	if m.ResolveExpenseAccountFunc != nil {
		return m.ResolveExpenseAccountFunc(ctx, variantID)
	}
	return 0, nil
}

type ConfigLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

func (m ConfigLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{}, Count: 0}, nil
}
