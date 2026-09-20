package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

type ApplyCreditRequest struct {
	OrganizationID uint64
	InvoiceID      uint64
	CreditNoteID   uint64
	Amount         float64
}

type ContraSettleRequest struct {
	OrganizationID    uint64
	CustomerInvoiceID uint64
	SupplierBillID    uint64
	JournalID         uint64
	Date              time.Time
	Amount            float64
}

type DeductDownPaymentRequest struct {
	OrganizationID   uint64
	FinalInvoiceID   uint64
	AdvanceInvoiceID uint64
	Amount           float64
}

func (s InvoiceService) WithSettlements(creditNotes InvoiceCreditApplicationDAO, contra InvoiceContraSettlementDAO) InvoiceService {
	s.creditApplications = creditNotes
	s.contraSettlements = contra
	return s
}

func (s InvoiceService) ListCreditApplications(ctx context.Context, invoiceID uint64) ([]*InvoiceCreditApplication, error) {
	if s.creditApplications == nil {
		return nil, nil
	}
	return s.creditApplications.ListByInvoice(ctx, invoiceID)
}

func (s InvoiceService) ListContraSettlements(ctx context.Context, invoiceID uint64) ([]*InvoiceContraSettlement, error) {
	if s.contraSettlements == nil {
		return nil, nil
	}
	return s.contraSettlements.ListByInvoice(ctx, invoiceID)
}

func (s InvoiceService) WithDownPayments(links DownPaymentLinkDAO) InvoiceService {
	s.downPayments = links
	return s
}

func (s InvoiceService) ListDownPayments(ctx context.Context, finalInvoiceID uint64) ([]*DownPaymentLink, error) {
	if s.downPayments == nil {
		return nil, nil
	}
	return s.downPayments.ListByFinal(ctx, finalInvoiceID)
}

func (s InvoiceService) DeductDownPayment(ctx context.Context, request DeductDownPaymentRequest) (*DownPaymentLink, error) {
	final, err := s.invoices.Find(ctx, request.FinalInvoiceID)
	if err != nil {
		return nil, err
	}
	advance, err := s.invoices.Find(ctx, request.AdvanceInvoiceID)
	if err != nil {
		return nil, err
	}
	if final == nil || final.OrganizationID == nil || *final.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if advance == nil || advance.OrganizationID == nil || *advance.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if final.ID == advance.ID || final.Type != advance.Type || !isPayableDocument(final.Type) {
		return nil, ErrCreditMismatch
	}
	if final.ContactID != advance.ContactID {
		return nil, ErrCreditMismatch
	}
	if final.State != InvoiceStatePosted || advance.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}
	finalResidual := final.AmountResidual.Round(4)
	advanceResidual := advance.AmountResidual.Round(4)
	if finalResidual.IsZero() || advanceResidual.IsZero() {
		return nil, ErrOverAllocation
	}
	deduct := amount.FromFloat64(request.Amount).Round(4)
	if deduct.IsZero() {
		deduct = finalResidual
		if advanceResidual.LessThan(deduct) {
			deduct = advanceResidual
		}
	}
	if !deduct.IsPositive() {
		return nil, ErrPaymentAmount
	}
	if deduct.GreaterThan(finalResidual) || deduct.GreaterThan(advanceResidual) {
		return nil, ErrOverAllocation
	}

	var link *DownPaymentLink
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		link = &DownPaymentLink{
			OrganizationID:   helper.Ptr(request.OrganizationID),
			AdvanceInvoiceID: advance.ID,
			FinalInvoiceID:   final.ID,
			Amount:           deduct,
		}
		var err error
		link, err = s.downPayments.Create(ctx, link)
		if err != nil {
			return err
		}
		final.AmountResidual = finalResidual.Sub(deduct).Round(4)
		final.PaymentState = residualState(final.AmountResidual)
		if _, err := s.invoices.UpdateTx(ctx, tx, final); err != nil {
			return err
		}
		advance.AmountResidual = advanceResidual.Sub(deduct).Round(4)
		advance.PaymentState = residualState(advance.AmountResidual)
		_, err = s.invoices.UpdateTx(ctx, tx, advance)
		return err
	})
	if err != nil {
		return nil, err
	}
	return link, nil
}

func isPayableDocument(invoiceType string) bool {
	return invoiceType == InvoiceTypeCustomerInvoice || invoiceType == InvoiceTypeSupplierBill
}

func (s InvoiceService) ApplyCredit(ctx context.Context, request ApplyCreditRequest) (*InvoiceCreditApplication, error) {
	invoice, err := s.invoices.Find(ctx, request.InvoiceID)
	if err != nil {
		return nil, err
	}
	credit, err := s.invoices.Find(ctx, request.CreditNoteID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil || *invoice.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if credit == nil || credit.OrganizationID == nil || *credit.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if !validCreditPair(invoice.Type, credit.Type) || invoice.ContactID != credit.ContactID {
		return nil, ErrCreditMismatch
	}
	if invoice.State != InvoiceStatePosted || credit.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}
	invoiceResidual := invoice.AmountResidual.Round(4)
	creditResidual := credit.AmountResidual.Round(4)
	if invoiceResidual.IsZero() || creditResidual.IsZero() {
		return nil, ErrOverAllocation
	}
	settle := amount.FromFloat64(request.Amount).Round(4)
	if settle.IsZero() {
		settle = invoiceResidual
		if creditResidual.LessThan(settle) {
			settle = creditResidual
		}
	}
	if !settle.IsPositive() {
		return nil, ErrPaymentAmount
	}
	if settle.GreaterThan(invoiceResidual) || settle.GreaterThan(creditResidual) {
		return nil, ErrOverAllocation
	}

	var applied *InvoiceCreditApplication
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		applied = &InvoiceCreditApplication{
			OrganizationID: helper.Ptr(request.OrganizationID),
			InvoiceID:      invoice.ID,
			CreditNoteID:   credit.ID,
			Amount:         settle,
		}
		var err error
		applied, err = s.creditApplications.Create(ctx, applied)
		if err != nil {
			return err
		}
		invoice.AmountResidual = invoiceResidual.Sub(settle).Round(4)
		invoice.PaymentState = residualState(invoice.AmountResidual)
		if _, err := s.invoices.UpdateTx(ctx, tx, invoice); err != nil {
			return err
		}
		credit.AmountResidual = creditResidual.Sub(settle).Round(4)
		credit.PaymentState = residualState(credit.AmountResidual)
		_, err = s.invoices.UpdateTx(ctx, tx, credit)
		return err
	})
	if err != nil {
		return nil, err
	}
	return applied, nil
}

func (s InvoiceService) ContraSettle(ctx context.Context, request ContraSettleRequest) (*InvoiceContraSettlement, error) {
	invoice, err := s.invoices.Find(ctx, request.CustomerInvoiceID)
	if err != nil {
		return nil, err
	}
	bill, err := s.invoices.Find(ctx, request.SupplierBillID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil || *invoice.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if bill == nil || bill.OrganizationID == nil || *bill.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if invoice.Type != InvoiceTypeCustomerInvoice || invoice.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}
	if bill.Type != InvoiceTypeSupplierBill || bill.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}
	if invoice.ContactID != bill.ContactID {
		return nil, ErrCreditMismatch
	}
	if !settlementCurrencyEqual(invoice.CurrencyCode, bill.CurrencyCode) {
		return nil, ErrCurrencyMismatch
	}
	invoiceResidual := invoice.AmountResidual.Round(4)
	billResidual := bill.AmountResidual.Round(4)
	if invoiceResidual.IsZero() || billResidual.IsZero() {
		return nil, ErrOverAllocation
	}
	settle := amount.FromFloat64(request.Amount).Round(4)
	if settle.IsZero() {
		settle = invoiceResidual
		if billResidual.LessThan(settle) {
			settle = billResidual
		}
	}
	if !settle.IsPositive() {
		return nil, ErrPaymentAmount
	}
	if settle.GreaterThan(invoiceResidual) || settle.GreaterThan(billResidual) {
		return nil, ErrOverAllocation
	}

	receivable, err := s.receivableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	payable, err := s.payableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	settleDate := request.Date
	if settleDate.IsZero() {
		settleDate = s.now().UTC()
	}
	functionalSettle, err := s.functional(ctx, request.OrganizationID, settle, invoice.CurrencyCode, settleDate)
	if err != nil {
		return nil, err
	}

	var settlement *InvoiceContraSettlement
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      request.JournalID,
			Date:           settleDate,
			Ref:            derefString(invoice.Name) + "/CONTRA",
			Description:    "AR/AP contra settlement",
			CurrencyCode:   invoice.CurrencyCode,
			Lines: []PostingLine{
				{AccountID: payable, Name: "Accounts Payable", Debit: functionalSettle, CurrencyCode: invoice.CurrencyCode, AmountCurrency: settle.Float64(), ContactID: helper.Ptr(invoice.ContactID)},
				{AccountID: receivable, Name: "Accounts Receivable", Credit: functionalSettle, CurrencyCode: invoice.CurrencyCode, AmountCurrency: settle.Float64(), ContactID: helper.Ptr(invoice.ContactID)},
			},
		})
		if err != nil {
			return err
		}
		settlement = &InvoiceContraSettlement{
			OrganizationID:    helper.Ptr(request.OrganizationID),
			CustomerInvoiceID: invoice.ID,
			SupplierBillID:    bill.ID,
			Amount:            settle,
			EntryID:           helper.Ptr(entry.ID),
			Date:              &settleDate,
		}
		settlement, err = s.contraSettlements.Create(ctx, settlement)
		if err != nil {
			return err
		}
		invoice.AmountResidual = invoiceResidual.Sub(settle).Round(4)
		invoice.PaymentState = residualState(invoice.AmountResidual)
		if _, err := s.invoices.UpdateTx(ctx, tx, invoice); err != nil {
			return err
		}
		bill.AmountResidual = billResidual.Sub(settle).Round(4)
		bill.PaymentState = residualState(bill.AmountResidual)
		_, err = s.invoices.UpdateTx(ctx, tx, bill)
		return err
	})
	if err != nil {
		return nil, err
	}
	return settlement, nil
}

func validCreditPair(invoiceType, creditType string) bool {
	return invoiceType == InvoiceTypeCustomerInvoice && creditType == InvoiceTypeCustomerCredit ||
		invoiceType == InvoiceTypeSupplierBill && creditType == InvoiceTypeSupplierCredit
}

func residualState(residual amount.Amount) string {
	if residual.IsZero() {
		return PaymentStatePaid
	}
	return PaymentStatePartial
}

func settlementCurrencyEqual(first, second *string) bool {
	if first == nil || second == nil {
		return first == second
	}
	return *first == *second
}
