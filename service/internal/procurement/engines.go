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
)

type OfferEngine interface {
	BestOfferForSupplier(ctx context.Context, variantID, supplierID uint64, qty amount.Amount, date time.Time) (*products.SupplierProduct, error)
}

type SupplierProfileLookup interface {
	FindByContact(ctx context.Context, contactID uint64) (*contacts.SupplierProfile, error)
}

type ReceiveEngine interface {
	Receive(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

type ApprovalEngine interface {
	IsApproved(ctx context.Context, ownerType string, ownerID uint64) (bool, error)
	Create(ctx context.Context, organizationID uint64, ownerType string, ownerID, requestedBy uint64, approverIDs []uint64) (*crosscutting.ApprovalRequest, error)
	State(ctx context.Context, ownerType string, ownerID uint64) (*crosscutting.ApprovalRequest, []*crosscutting.ApprovalStep, error)
}

type BillEngine interface {
	CreateSupplierBill(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error)
	CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}

type OpenInvoiceLookup interface {
	ListOpenByContact(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error)
}

type OutboundPaymentEngine interface {
	CreateOutbound(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error)
}

type QualityEngine interface {
	TriggerChecks(ctx context.Context, organizationID uint64, shipmentID uint64, itemIDs []uint64) (int, error)
	HasFailedChecks(ctx context.Context, shipmentID uint64) (bool, error)
}

type CurrencyConverter interface {
	Convert(ctx context.Context, amount amount.Amount, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (amount.Amount, error)
	HasRate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (bool, error)
}

type CreateFromAgreementEngine interface {
	CreateFromAgreement(ctx context.Context, agreementID uint64, overrides *CreateFromAgreementOverrides) (*PurchaseOrder, error)
}

type PaymentBatchService interface {
	CreatePaymentBatch(ctx context.Context, organizationID, supplierID uint64, request PaymentBatchRequest) (*PaymentBatch, []*PaymentBatchLine, error)
	ConfirmBatch(ctx context.Context, batchID uint64) (*PaymentBatch, error)
	GetPaymentBatch(ctx context.Context, batchID uint64) (*PaymentBatch, []*PaymentBatchLine, error)
	ListPaymentBatches(ctx context.Context, q *query.Query) (*query.Page[PaymentBatch], error)
}
