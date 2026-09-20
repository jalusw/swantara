package interorganization

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

type SupplierBillCreator interface {
	CreateSupplierBill(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error)
}

func (s InterorganizationService) WithInvoiceMirror(invoices accounting.InvoiceDAO, lines accounting.InvoiceLineDAO, bills SupplierBillCreator) InterorganizationService {
	s.invInvoices = invoices
	s.invLines = lines
	s.invBills = bills
	return s
}

type MirrorInvoiceRequest struct {
	InvoiceID        uint64
	ToOrganizationID uint64
	JournalID        uint64
	ExpenseAccountID uint64
}

func (s InterorganizationService) MirrorInvoice(ctx context.Context, request MirrorInvoiceRequest) (*accounting.Invoice, *InterorganizationTransaction, error) {
	if s.invInvoices == nil || s.invBills == nil {
		return nil, nil, ErrMirrorRuleMissing
	}
	invoice, err := s.invInvoices.Find(ctx, request.InvoiceID)
	if err != nil {
		return nil, nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil {
		return nil, nil, ErrMirrorSourceNotFound
	}
	if invoice.Type != accounting.InvoiceTypeCustomerInvoice || invoice.State != accounting.InvoiceStatePosted {
		return nil, nil, ErrMirrorSourceNotPosted
	}

	rule, err := s.ruleFor(ctx, *invoice.OrganizationID, request.ToOrganizationID)
	if err != nil {
		return nil, nil, err
	}
	if rule == nil {
		return nil, nil, ErrMirrorRuleMissing
	}
	if !rule.AutoMirror {
		return nil, nil, ErrMirrorRuleDisabled
	}
	if rule.SupplierContactID == nil {
		return nil, nil, ErrRuleContactsRequired
	}

	sourceLines, err := s.invLines.ListByInvoice(ctx, invoice.ID)
	if err != nil {
		return nil, nil, err
	}
	lines := make([]accounting.InvoiceLineRequest, 0, len(sourceLines))
	for _, line := range sourceLines {
		if line.Qty <= 0 {
			continue
		}
		lines = append(lines, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: helper.Deref(line.Description, ""),
			Qty:         line.Qty,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice.Float64(),
			DiscountPct: line.DiscountPct,
			AccountID:   request.ExpenseAccountID,
		})
	}
	if len(lines) == 0 {
		return nil, nil, ErrMirrorSourceNoLines
	}

	invoiceDate := time.Now().UTC()
	if invoice.InvoiceDate != nil {
		invoiceDate = *invoice.InvoiceDate
	}
	bill, err := s.invBills.CreateSupplierBill(ctx, accounting.CreateSupplierBillRequest{
		OrganizationID: request.ToOrganizationID,
		JournalID:      request.JournalID,
		ContactID:      *rule.SupplierContactID,
		Date:           invoiceDate,
		DueDate:        invoice.DueDate,
		Reference:      helper.Deref(invoice.Name, ""),
		CurrencyCode:   invoice.CurrencyCode,
		Lines:          lines,
	})
	if err != nil {
		return nil, nil, err
	}

	transaction, err := s.trans.Create(ctx, &InterorganizationTransaction{
		SourceOrganizationID: invoice.OrganizationID,
		SourceType:           "customer_invoice",
		SourceID:             helper.Ptr(invoice.ID),
		MirrorOrganizationID: helper.Ptr(request.ToOrganizationID),
		MirrorType:           "invoice",
		MirrorID:             helper.Ptr(bill.ID),
		Amount:               bill.AmountTotal.Float64(),
		State:                TransactionStateDone,
	})
	if err != nil {
		return nil, nil, err
	}
	return bill, transaction, nil
}
