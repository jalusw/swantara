package service

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

type InvoiceBuilder interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}
