package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type IncomeAccountResolver interface {
	ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error)
}

type InvoiceEngine interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

type ProjectLookup interface {
	Search(ctx context.Context, field string, value any) (*project.Project, error)
}
