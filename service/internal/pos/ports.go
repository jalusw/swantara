package pos

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type StockValuer interface {
	Ship(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	ShipTx(ctx context.Context, tx *gorm.DB, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	Restock(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	RestockTx(ctx context.Context, tx *gorm.DB, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

type InvoiceEngine interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
	CreateTx(ctx context.Context, tx *gorm.DB, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
	CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
	CreateCreditNoteTx(ctx context.Context, tx *gorm.DB, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}

type AccountLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

type PaymentAccountLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.POSPaymentAccount], error)
}
