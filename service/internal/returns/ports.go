package returns

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

type OriginOrderLookup interface {
	CustomerOrder(ctx context.Context, orderID uint64) (uint64, error)
	VendorOrder(ctx context.Context, orderID uint64) (uint64, error)
}

type SaleOrderLookup interface {
	Find(ctx context.Context, id uint64) (*sales.SaleOrder, error)
}

type PurchaseOrderLookup interface {
	Find(ctx context.Context, id uint64) (*procurement.PurchaseOrder, error)
}

type SaleOrderReturnRecorder interface {
	RecordReturn(ctx context.Context, orderID uint64, itemID uint64, qty float64) error
}

type PurchaseOrderReturnRecorder interface {
	RecordReturn(ctx context.Context, orderID uint64, itemID uint64, qty float64) error
}

type StockMovementLookup interface {
	ListByOrigin(ctx context.Context, originType string, originID uint64) ([]*inventory.StockMovement, error)
	Create(ctx context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error)
	FindForUpdateTx(ctx context.Context, tx *gorm.DB, moveID uint64) (*inventory.StockMovement, error)
	ApplyTx(ctx context.Context, tx *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error)
}

type StockLocationLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error)
}

type StockLayerLookup interface {
	ListByMovement(ctx context.Context, moveID uint64) ([]*inventory.CostLayer, error)
}

type ReturnValuer interface {
	Restock(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	ReturnToSupplier(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

type InvoiceLookup interface {
	FindBySaleOrder(ctx context.Context, orderID uint64) (*accounting.Invoice, error)
	FindByPurchaseOrder(ctx context.Context, orderID uint64) (*accounting.Invoice, error)
}

type CreditNoteEngine interface {
	CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
	CreateVendorCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}
