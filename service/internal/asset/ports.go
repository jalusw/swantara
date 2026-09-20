package asset

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type AssetCategoryLookup interface {
	Find(ctx context.Context, id uint64) (*reference.AssetCategory, error)
}

type InvoiceLineLookup interface {
	Find(ctx context.Context, id uint64) (*accounting.InvoiceLine, error)
}

type InvoiceLookup interface {
	Find(ctx context.Context, id uint64) (*accounting.Invoice, error)
}
