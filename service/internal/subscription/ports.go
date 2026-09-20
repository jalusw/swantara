package subscription

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"gorm.io/gorm"
)

type IncomeAccountResolver interface {
	ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error)
	ResolveVariantOrganization(ctx context.Context, variantID uint64) (*uint64, error)
}

type InvoiceEngine interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
	CreateTx(ctx context.Context, tx *gorm.DB, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}
