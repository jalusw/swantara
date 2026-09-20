package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

type InvoiceHandler struct {
	svc accounting.InvoiceService
}

func NewInvoiceHandler(svc accounting.InvoiceService) InvoiceHandler {
	return InvoiceHandler{svc: svc}
}

func (h InvoiceHandler) buildLines(requestLines []InvoiceLineRequest) []accounting.InvoiceLineRequest {
	lines := make([]accounting.InvoiceLineRequest, len(requestLines))
	for i, line := range requestLines {
		lines[i] = accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: line.Description,
			Qty:         line.Qty,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			AccountID:   line.AccountID,
			DimensionID: line.DimensionID,
		}
	}
	return lines
}

func (h InvoiceHandler) parseDates(request CreateInvoiceRequest) (*time.Time, *time.Time, error) {
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return nil, nil, err
	}
	dueDate, err := helper.ParseDate(request.DueDate)
	if err != nil {
		return nil, nil, err
	}
	return date, dueDate, nil
}
