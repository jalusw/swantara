package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func (s InvoiceService) Post(ctx context.Context, invoiceID, organizationID uint64) (*Invoice, error) {
	invoice, err := s.invoices.Find(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil || *invoice.OrganizationID != organizationID {
		return nil, ErrInvoiceNotFound
	}
	if invoice.State != InvoiceStateDraft {
		return nil, ErrInvoiceNotDraft
	}
	if invoice.JournalID == nil {
		return nil, ErrClosingJournalMissing
	}

	persisted, err := s.lines.ListByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	requestLines := make([]InvoiceLineRequest, 0, len(persisted))
	for _, line := range persisted {
		if line.AccountID == nil {
			return nil, ErrNoRevenueAccount
		}
		description := ""
		if line.Description != nil {
			description = *line.Description
		}
		requestLines = append(requestLines, InvoiceLineRequest{
			ItemID:         line.ItemID,
			Description:    description,
			Qty:            line.Qty,
			UnitID:         line.UnitID,
			UnitPrice:      line.UnitPrice.Float64(),
			DiscountPct:    line.DiscountPct,
			TaxIDs:         line.TaxIDs,
			AccountID:      *line.AccountID,
			DimensionID:    line.DimensionID,
			SaleLineID:     line.SaleLineID,
			PurchaseLineID: line.PurchaseLineID,
		})
	}

	scope := reference.TaxScopeSale
	origin := InvoiceTypeCustomerInvoice
	postDescription := "Customer invoice"
	if invoice.Type == InvoiceTypeSupplierBill {
		scope = reference.TaxScopePurchase
		origin = InvoiceTypeSupplierBill
		postDescription = "Supplier bill"
	}
	computed, err := s.computeInvoice(ctx, organizationID, requestLines, scope)
	if err != nil {
		return nil, err
	}

	invoiceDate := s.now().UTC()
	if invoice.InvoiceDate != nil && !invoice.InvoiceDate.IsZero() {
		invoiceDate = *invoice.InvoiceDate
	}

	var postLines []PostingLine
	if invoice.Type == InvoiceTypeSupplierBill {
		payable, err := s.payableAccount(ctx, organizationID)
		if err != nil {
			return nil, err
		}
		postLines, err = s.supplierPostLines(ctx, organizationID, invoice.ContactID, invoice.CurrencyCode, invoice.DueDate, invoiceDate, payable, computed)
		if err != nil {
			return nil, err
		}
	} else {
		postLines, err = s.customerPostLines(ctx, organizationID, invoice.ContactID, invoice.CurrencyCode, invoice.DueDate, nil, invoiceDate, computed)
		if err != nil {
			return nil, err
		}
	}

	var posted *Invoice
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      *invoice.JournalID,
			Date:           invoiceDate,
			Ref:            derefString(invoice.Name),
			OriginType:     origin,
			OriginID:       invoice.ID,
			Description:    postDescription,
			CurrencyCode:   invoice.CurrencyCode,
			Lines:          postLines,
		})
		if err != nil {
			return err
		}
		invoice.EntryID = helper.Ptr(entry.ID)
		invoice.State = InvoiceStatePosted
		posted, err = s.invoices.UpdateTx(ctx, tx, invoice)
		return err
	})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

type WriteOffBadDebtRequest struct {
	InvoiceID        uint64
	OrganizationID   uint64
	ExpenseAccountID uint64
	Date             time.Time
}

func (s InvoiceService) WriteOffBadDebt(ctx context.Context, request WriteOffBadDebtRequest) (*Invoice, error) {
	invoice, err := s.invoices.Find(ctx, request.InvoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil || *invoice.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if invoice.Type != InvoiceTypeCustomerInvoice || invoice.State != InvoiceStatePosted || !invoice.AmountResidual.IsPositive() {
		return nil, ErrBadDebtNotAllowed
	}
	if invoice.JournalID == nil {
		return nil, ErrClosingJournalMissing
	}

	receivable, err := s.receivableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}

	writeOffDate := request.Date
	if writeOffDate.IsZero() {
		writeOffDate = s.now().UTC()
	}
	functionalResidual, err := s.functional(ctx, request.OrganizationID, invoice.AmountResidual, invoice.CurrencyCode, writeOffDate)
	if err != nil {
		return nil, err
	}

	var writtenOff *Invoice
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      derefUint64(invoice.JournalID),
			Date:           writeOffDate,
			Ref:            derefString(invoice.Name) + "/WO",
			OriginType:     OriginTypeBadDebt,
			OriginID:       invoice.ID,
			Description:    "Bad debt write-off",
			CurrencyCode:   invoice.CurrencyCode,
			Lines: []PostingLine{
				{AccountID: request.ExpenseAccountID, Name: "Bad Debt Expense", Debit: functionalResidual, CurrencyCode: invoice.CurrencyCode, AmountCurrency: invoice.AmountResidual.Float64(), ContactID: helper.Ptr(invoice.ContactID)},
				{AccountID: receivable, Name: "Accounts Receivable", Credit: functionalResidual, CurrencyCode: invoice.CurrencyCode, AmountCurrency: invoice.AmountResidual.Float64(), ContactID: helper.Ptr(invoice.ContactID)},
			},
		})
		if err != nil {
			return err
		}
		invoice.AmountResidual = amount.Zero()
		invoice.PaymentState = PaymentStateBadDebt
		writtenOff, err = s.invoices.UpdateTx(ctx, tx, invoice)
		return err
	})
	if err != nil {
		return nil, err
	}
	return writtenOff, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefUint64(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}
