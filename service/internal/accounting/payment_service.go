package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type CreatePaymentRequest struct {
	OrganizationID    uint64
	ContactID         uint64
	JournalID         uint64
	Amount            float64
	CurrencyCode      *string
	Date              time.Time
	Reference         string
	InvoiceIDs        []uint64
	Tolerance         float64
	AllowAdvance      bool
	DiscountAmount    float64
	WithholdingAmount float64
	WriteOffAccountID *uint64
}

type PaymentService struct {
	payments    PaymentDAO
	allocations PaymentAllocationDAO
	invoices    InvoiceDAO
	poster      Poster
	accounts    AccountLookup
	journals    dao.CRUD[reference.Journal]
	sequences   sequence.Service
	tx          db.Transactioner
	now         func() time.Time
	rates       CurrencyRateResolver
	fx          FxAccountResolver
	orgs        OrganizationLookup
	advances    AdvanceAccountResolver
}

func NewPaymentService(
	payments PaymentDAO,
	invoices InvoiceDAO,
	poster Poster,
	accounts AccountLookup,
	journals dao.CRUD[reference.Journal],
	sequences sequence.Service,
	tx db.Transactioner,
) PaymentService {
	return PaymentService{
		payments:  payments,
		invoices:  invoices,
		poster:    poster,
		accounts:  accounts,
		journals:  journals,
		sequences: sequences,
		tx:        tx,
		now:       time.Now,
	}
}

func (s PaymentService) WithFxResolvers(rates CurrencyRateResolver, fx FxAccountResolver) PaymentService {
	s.rates = rates
	s.fx = fx
	return s
}

func (s PaymentService) WithOrganizations(orgs OrganizationLookup) PaymentService {
	s.orgs = orgs
	return s
}

func (s PaymentService) WithAdvanceAccounts(advances AdvanceAccountResolver) PaymentService {
	s.advances = advances
	return s
}

func (s PaymentService) WithAllocations(allocations PaymentAllocationDAO) PaymentService {
	s.allocations = allocations
	return s
}

func (s PaymentService) Find(ctx context.Context, id uint64) (*Payment, error) {
	return s.payments.Find(ctx, id)
}

func (s PaymentService) List(ctx context.Context, q *query.Query) (*query.Page[Payment], error) {
	return s.payments.List(ctx, q)
}

func (s PaymentService) ListAllocations(ctx context.Context, paymentID uint64) ([]*PaymentAllocation, error) {
	return s.allocations.ListByPayment(ctx, paymentID)
}

func (s PaymentService) Create(ctx context.Context, request CreatePaymentRequest) (*Payment, error) {
	return s.createInbound(ctx, nil, request)
}

func (s PaymentService) CreateTx(ctx context.Context, tx *gorm.DB, request CreatePaymentRequest) (*Payment, error) {
	return s.createInbound(ctx, tx, request)
}

func (s PaymentService) createInbound(ctx context.Context, tx *gorm.DB, request CreatePaymentRequest) (*Payment, error) {
	if !amount.FromFloat64(request.Amount).GreaterThan(amount.Zero()) {
		return nil, ErrPaymentAmount
	}
	if len(request.InvoiceIDs) == 0 {
		return nil, ErrPaymentNoInvoices
	}

	bankAccount, err := s.bankAccount(ctx, request.JournalID)
	if err != nil {
		return nil, err
	}
	receivable, err := s.receivableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	if (request.DiscountAmount != 0 || request.WithholdingAmount != 0) && request.WriteOffAccountID == nil {
		return nil, ErrReconcileWriteOff
	}
	effectiveAmount := request.Amount + request.DiscountAmount + request.WithholdingAmount
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	preOpen, err := s.loadOpen(ctx, request.InvoiceIDs)
	if err != nil {
		return nil, err
	}
	if s.crossCurrency(preOpen, request.CurrencyCode) {
		return s.createCrossCurrency(ctx, tx, request, bankAccount, receivable, date, true)
	}
	open, allocations, openResidual, err := s.allocate(ctx, request.InvoiceIDs, effectiveAmount, request.Tolerance, request.AllowAdvance)
	if err != nil {
		return nil, err
	}
	for _, inv := range open {
		if !currencyEqual(inv.CurrencyCode, request.CurrencyCode) {
			return nil, ErrCurrencyMismatch
		}
	}

	name, err := s.sequences.Next(ctx, request.OrganizationID, SequencePaymentCode)
	if err != nil {
		return nil, err
	}

	payment := &Payment{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           helper.Ptr(name),
		ContactID:      request.ContactID,
		Type:           PaymentTypeInbound,
		JournalID:      helper.Ptr(request.JournalID),
		Amount:         request.Amount,
		CurrencyCode:   request.CurrencyCode,
		Date:           date,
		Reference:      helper.Ptr(request.Reference),
		State:          PaymentStatePosted,
	}

	postLines := []PostingLine{
		{AccountID: bankAccount, Name: "Bank", Debit: amount.FromFloat64(request.Amount), CurrencyCode: request.CurrencyCode, AmountCurrency: request.Amount},
		{AccountID: receivable, Name: "Accounts Receivable", Credit: amount.FromFloat64(effectiveAmount), CurrencyCode: request.CurrencyCode, AmountCurrency: effectiveAmount},
	}
	over := amount.FromFloat64(effectiveAmount).Sub(openResidual).Round(4)
	if over.IsPositive() {
		advanceAccount, err := s.advanceAccountID(ctx, request.OrganizationID, true)
		if err != nil {
			return nil, err
		}
		postLines[1].Credit = amount.FromFloat64(effectiveAmount).Sub(over).Round(4)
		postLines[1].AmountCurrency = amount.FromFloat64(effectiveAmount).Sub(over).Float64()
		postLines = append(postLines, PostingLine{AccountID: advanceAccount, Name: "Customer Advance", Credit: over})
	}
	if request.DiscountAmount != 0 && request.WriteOffAccountID != nil {
		postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Cash Discount", Debit: amount.FromFloat64(request.DiscountAmount)})
	}
	if request.WithholdingAmount != 0 && request.WriteOffAccountID != nil {
		postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Withholding", Debit: amount.FromFloat64(request.WithholdingAmount)})
	}
	if s.rates != nil && s.fx != nil && request.CurrencyCode != nil {
		fxLines, err := s.realizedFxLines(ctx, request.OrganizationID, request.CurrencyCode, date, open, allocations, true)
		if err != nil {
			return nil, err
		}
		if len(fxLines) > 0 {
			functionalBank := amount.Zero()
			functionalReceivable := amount.Zero()
			for i, alloc := range allocations {
				foreign := amount.FromFloat64(alloc.Amount)
				inv := open[i]
				if inv.CurrencyCode != nil && *inv.CurrencyCode == *request.CurrencyCode {
					invRate, err := s.rateFor(ctx, request.OrganizationID, *inv.CurrencyCode, inv.InvoiceDate, date)
					if err != nil {
						return nil, err
					}
					payRate, err := s.rates.Rate(ctx, *request.CurrencyCode, request.OrganizationID, amount.RateSpot, date)
					if err != nil {
						return nil, err
					}
					functionalBank = functionalBank.Add(foreign.Mul(payRate).Round(4))
					functionalReceivable = functionalReceivable.Add(foreign.Mul(invRate).Round(4))
				} else {
					functionalBank = functionalBank.Add(foreign)
					functionalReceivable = functionalReceivable.Add(foreign)
				}
			}
			if !functionalBank.IsZero() {
				postLines[0].Debit = functionalBank
			}
			if !functionalReceivable.IsZero() {
				postLines[1].Credit = functionalReceivable
			}
			postLines = append(postLines, fxLines...)
		}
	}
	return s.runPayment(ctx, tx, txPayment{payment: payment, date: date, name: name, journalID: request.JournalID, organizationID: request.OrganizationID, description: "Customer payment", open: open, allocations: allocations, postLines: postLines})
}

func (s PaymentService) CreateOutbound(ctx context.Context, request CreatePaymentRequest) (*Payment, error) {
	return s.createOutbound(ctx, nil, request)
}

func (s PaymentService) CreateOutboundTx(ctx context.Context, tx *gorm.DB, request CreatePaymentRequest) (*Payment, error) {
	return s.createOutbound(ctx, tx, request)
}

func (s PaymentService) createOutbound(ctx context.Context, tx *gorm.DB, request CreatePaymentRequest) (*Payment, error) {
	if !amount.FromFloat64(request.Amount).GreaterThan(amount.Zero()) {
		return nil, ErrPaymentAmount
	}
	if len(request.InvoiceIDs) == 0 {
		return nil, ErrPaymentNoInvoices
	}

	bankAccount, err := s.bankAccount(ctx, request.JournalID)
	if err != nil {
		return nil, err
	}
	payable, err := s.payableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	if (request.DiscountAmount != 0 || request.WithholdingAmount != 0) && request.WriteOffAccountID == nil {
		return nil, ErrReconcileWriteOff
	}
	effectiveAmount := request.Amount + request.DiscountAmount + request.WithholdingAmount

	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	preOpen, err := s.loadOpen(ctx, request.InvoiceIDs)
	if err != nil {
		return nil, err
	}
	if s.crossCurrency(preOpen, request.CurrencyCode) {
		return s.createCrossCurrency(ctx, tx, request, payable, bankAccount, date, false)
	}
	open, allocations, openResidual, err := s.allocate(ctx, request.InvoiceIDs, effectiveAmount, request.Tolerance, request.AllowAdvance)
	if err != nil {
		return nil, err
	}
	for _, inv := range open {
		if !currencyEqual(inv.CurrencyCode, request.CurrencyCode) {
			return nil, ErrCurrencyMismatch
		}
	}

	name, err := s.sequences.Next(ctx, request.OrganizationID, SequencePaymentCode)
	if err != nil {
		return nil, err
	}

	payment := &Payment{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           helper.Ptr(name),
		ContactID:      request.ContactID,
		Type:           PaymentTypeOutbound,
		JournalID:      helper.Ptr(request.JournalID),
		Amount:         request.Amount,
		CurrencyCode:   request.CurrencyCode,
		Date:           date,
		Reference:      helper.Ptr(request.Reference),
		State:          PaymentStatePosted,
	}

	postLines := []PostingLine{
		{AccountID: payable, Name: "Accounts Payable", Debit: amount.FromFloat64(effectiveAmount), CurrencyCode: request.CurrencyCode, AmountCurrency: effectiveAmount},
		{AccountID: bankAccount, Name: "Bank", Credit: amount.FromFloat64(request.Amount), CurrencyCode: request.CurrencyCode, AmountCurrency: request.Amount},
	}
	if request.DiscountAmount != 0 && request.WriteOffAccountID != nil {
		postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Cash Discount", Credit: amount.FromFloat64(request.DiscountAmount)})
	}
	if request.WithholdingAmount != 0 && request.WriteOffAccountID != nil {
		postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Withholding", Credit: amount.FromFloat64(request.WithholdingAmount)})
	}
	over := amount.FromFloat64(effectiveAmount).Sub(openResidual).Round(4)
	if over.IsPositive() {
		advanceAccount, err := s.advanceAccountID(ctx, request.OrganizationID, false)
		if err != nil {
			return nil, err
		}
		postLines[0].Debit = amount.FromFloat64(effectiveAmount).Sub(over).Round(4)
		postLines = append(postLines, PostingLine{AccountID: advanceAccount, Name: "Supplier Advance", Debit: over})
	}
	if s.rates != nil && s.fx != nil && request.CurrencyCode != nil {
		fxLines, err := s.realizedFxLines(ctx, request.OrganizationID, request.CurrencyCode, date, open, allocations, false)
		if err != nil {
			return nil, err
		}
		if len(fxLines) > 0 {
			functionalPayable := amount.Zero()
			functionalBank := amount.Zero()
			for i, alloc := range allocations {
				foreign := amount.FromFloat64(alloc.Amount)
				inv := open[i]
				if inv.CurrencyCode != nil && *inv.CurrencyCode == *request.CurrencyCode {
					invRate, err := s.rateFor(ctx, request.OrganizationID, *inv.CurrencyCode, inv.InvoiceDate, date)
					if err != nil {
						return nil, err
					}
					payRate, err := s.rates.Rate(ctx, *request.CurrencyCode, request.OrganizationID, amount.RateSpot, date)
					if err != nil {
						return nil, err
					}
					functionalPayable = functionalPayable.Add(foreign.Mul(invRate).Round(4))
					functionalBank = functionalBank.Add(foreign.Mul(payRate).Round(4))
				} else {
					functionalPayable = functionalPayable.Add(foreign)
					functionalBank = functionalBank.Add(foreign)
				}
			}
			if !functionalPayable.IsZero() {
				postLines[0].Debit = functionalPayable
			}
			if !functionalBank.IsZero() {
				postLines[1].Credit = functionalBank
			}
			postLines = append(postLines, fxLines...)
		}
	}
	return s.runPayment(ctx, tx, txPayment{payment: payment, date: date, name: name, journalID: request.JournalID, organizationID: request.OrganizationID, description: "Supplier payment", open: open, allocations: allocations, postLines: postLines})
}

type txPayment struct {
	payment        *Payment
	date           time.Time
	name           string
	journalID      uint64
	organizationID uint64
	description    string
	open           []*Invoice
	allocations    []*PaymentAllocation
	postLines      []PostingLine
}

func (s PaymentService) runPayment(ctx context.Context, tx *gorm.DB, txInput txPayment) (*Payment, error) {
	var created *Payment
	apply := func(tx *gorm.DB) error {
		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: txInput.organizationID,
			JournalID:      txInput.journalID,
			Date:           txInput.date,
			Ref:            txInput.name,
			OriginType:     txInput.payment.Type,
			Description:    txInput.description,
			Lines:          txInput.postLines,
		})
		if err != nil {
			return err
		}
		txInput.payment.EntryID = helper.Ptr(entry.ID)

		created, err = s.payments.CreateWithAllocationsTx(ctx, tx, txInput.payment, txInput.allocations)
		if err != nil {
			return err
		}
		for _, invoice := range txInput.open {
			if _, err := s.invoices.UpdateTx(ctx, tx, invoice); err != nil {
				return err
			}
		}
		return nil
	}
	var err error
	if tx != nil {
		err = apply(tx)
	} else {
		err = s.tx.Run(ctx, apply)
	}
	if err != nil {
		return nil, err
	}
	created.State = PaymentStateReconciled
	return created, nil
}

func (s PaymentService) allocate(ctx context.Context, invoiceIDs []uint64, amountToPay float64, tolerance float64, allowAdvance bool) ([]*Invoice, []*PaymentAllocation, amount.Amount, error) {
	open := make([]*Invoice, 0, len(invoiceIDs))
	openResidual := amount.Zero()
	for _, invoiceID := range invoiceIDs {
		invoice, err := s.invoices.Find(ctx, invoiceID)
		if err != nil {
			return nil, nil, amount.Amount{}, err
		}
		if invoice == nil {
			return nil, nil, amount.Amount{}, ErrInvoiceNotFound
		}
		if invoice.State != InvoiceStatePosted {
			return nil, nil, amount.Amount{}, ErrInvoiceNotPosted
		}
		residual := invoice.AmountResidual.Round(4)
		if residual.IsZero() {
			continue
		}
		open = append(open, invoice)
		openResidual = openResidual.Add(residual)
	}
	over := amount.FromFloat64(amountToPay).Sub(openResidual)
	if over.IsPositive() {
		tol := amount.FromFloat64(tolerance).Abs().Round(4)
		if tol.IsZero() {
			tol = amount.FromFloat64(0.01).Round(4)
		}
		if over.GreaterThan(tol) && !allowAdvance {
			return nil, nil, amount.Amount{}, ErrOverAllocation
		}
	}

	payable := amount.FromFloat64(amountToPay).Round(4)
	allocations := make([]*PaymentAllocation, 0, len(open))
	for _, invoice := range open {
		residual := invoice.AmountResidual.Round(4)
		allocated := residual
		if payable.LessThan(residual) {
			allocated = payable
		}
		invoice.AmountResidual = residual.Sub(allocated).Round(4)
		if invoice.AmountResidual.IsZero() {
			invoice.PaymentState = PaymentStatePaid
		} else {
			invoice.PaymentState = PaymentStatePartial
		}
		allocations = append(allocations, &PaymentAllocation{InvoiceID: invoice.ID, Amount: allocated.Float64()})
		payable = payable.Sub(allocated)
		if payable.IsZero() {
			break
		}
	}
	return open, allocations, openResidual, nil
}

func (s PaymentService) realizedFxLines(ctx context.Context, organizationID uint64, paymentCurrency *string, paymentDate time.Time, open []*Invoice, allocations []*PaymentAllocation, inbound bool) ([]PostingLine, error) {
	if s.rates == nil || s.fx == nil || paymentCurrency == nil {
		return nil, nil
	}
	var lines []PostingLine
	for i, inv := range open {
		if i >= len(allocations) {
			continue
		}
		if inv.CurrencyCode == nil || *inv.CurrencyCode != *paymentCurrency {
			continue
		}
		foreign := amount.FromFloat64(allocations[i].Amount)
		if foreign.IsZero() {
			continue
		}
		invRate, err := s.rateFor(ctx, organizationID, *inv.CurrencyCode, inv.InvoiceDate, paymentDate)
		if err != nil {
			return nil, err
		}
		payRate, err := s.rates.Rate(ctx, *paymentCurrency, organizationID, amount.RateSpot, paymentDate)
		if err != nil {
			return nil, err
		}
		if payRate.Equal(invRate) {
			continue
		}
		diff := foreign.Mul(payRate.Sub(invRate)).Round(4)
		if diff.IsZero() {
			continue
		}
		if !inbound {
			diff = diff.Neg()
		}
		if diff.IsPositive() {
			gainID, err := s.fx.FxGainAccountID(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			if gainID == 0 {
				continue
			}
			lines = append(lines, PostingLine{AccountID: gainID, Name: "FX Gain", Credit: diff})
		} else {
			lossID, err := s.fx.FxLossAccountID(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			if lossID == 0 {
				continue
			}
			lines = append(lines, PostingLine{AccountID: lossID, Name: "FX Loss", Debit: diff.Abs()})
		}
	}
	if len(lines) == 0 {
		return nil, nil
	}
	merged := map[uint64]PostingLine{}
	for _, line := range lines {
		if existing, ok := merged[line.AccountID]; ok {
			existing.Debit = existing.Debit.Add(line.Debit)
			existing.Credit = existing.Credit.Add(line.Credit)
			merged[line.AccountID] = existing
		} else {
			merged[line.AccountID] = line
		}
	}
	out := make([]PostingLine, 0, len(merged))
	for _, line := range merged {
		if line.Debit.IsZero() && line.Credit.IsZero() {
			continue
		}
		out = append(out, line)
	}
	return out, nil
}

func (s PaymentService) rateFor(ctx context.Context, organizationID uint64, currencyCode string, invoiceDate *time.Time, fallbackDate time.Time) (amount.Amount, error) {
	date := fallbackDate
	if invoiceDate != nil && !invoiceDate.IsZero() {
		date = *invoiceDate
	}
	rate, err := s.rates.Rate(ctx, currencyCode, organizationID, amount.RateSpot, date)
	if err != nil {
		return amount.Amount{}, err
	}
	return rate, nil
}

func (s PaymentService) bankAccount(ctx context.Context, journalID uint64) (uint64, error) {
	journal, err := s.journals.Find(ctx, journalID)
	if err != nil {
		return 0, err
	}
	if journal == nil {
		return 0, ErrEntryNotFound
	}
	if journal.DefaultAccountID == nil {
		return 0, ErrNoBankAccount
	}
	return *journal.DefaultAccountID, nil
}

func (s PaymentService) receivableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: "receivable"}}})
	if err != nil {
		return 0, err
	}
	for _, account := range page.Items {
		if account.OrganizationID == organizationID && account.Active {
			return account.ID, nil
		}
	}
	return 0, ErrNoReceivableAccount
}

func (s PaymentService) Void(ctx context.Context, paymentID, organizationID uint64) (*Payment, error) {
	payment, err := s.payments.Find(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil || payment.OrganizationID == nil || *payment.OrganizationID != organizationID {
		return nil, ErrPaymentNotFound
	}
	if payment.State == PaymentStateCancelled {
		return nil, ErrEntryReversed
	}
	if payment.EntryID == nil {
		return nil, ErrEntryNotFound
	}
	if payment.JournalID == nil {
		return nil, ErrNoBankAccount
	}
	var voided *Payment
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err = s.poster.ReverseTx(ctx, tx, ReverseRequest{
			OrganizationID: organizationID,
			JournalID:      *payment.JournalID,
			Date:           s.now().UTC(),
			Ref:            *payment.Name + "/VOID",
			Description:    "Void payment",
			EntryID:        *payment.EntryID,
		})
		if err != nil {
			return err
		}
		if s.allocations != nil {
			allocations, err := s.allocations.ListByPayment(ctx, payment.ID)
			if err != nil {
				return err
			}
			for _, alloc := range allocations {
				invoice, err := s.invoices.Find(ctx, alloc.InvoiceID)
				if err != nil {
					return err
				}
				if invoice == nil {
					continue
				}
				invoice.AmountResidual = invoice.AmountResidual.Add(amount.FromFloat64(alloc.Amount)).Round(4)
				if invoice.AmountResidual.GreaterThan(amount.Zero()) && invoice.AmountResidual.LessThan(invoice.AmountTotal.Round(4)) {
					invoice.PaymentState = PaymentStatePartial
				} else {
					invoice.PaymentState = PaymentStateNotPaid
				}
				if _, err := s.invoices.UpdateTx(ctx, tx, invoice); err != nil {
					return err
				}
				if err := s.allocations.DeleteTx(ctx, tx, alloc.ID); err != nil {
					return err
				}
			}
		}
		payment.State = PaymentStateCancelled
		updated, err := s.payments.UpdateTx(ctx, tx, payment)
		if err != nil {
			return err
		}
		voided = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return voided, nil
}

func (s PaymentService) payableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: "payable"}}})
	if err != nil {
		return 0, err
	}
	for _, account := range page.Items {
		if account.OrganizationID == organizationID && account.Active {
			return account.ID, nil
		}
	}
	return 0, ErrNoPayableAccount
}

func currencyEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (s PaymentService) loadOpen(ctx context.Context, invoiceIDs []uint64) ([]*Invoice, error) {
	open := make([]*Invoice, 0, len(invoiceIDs))
	for _, id := range invoiceIDs {
		invoice, err := s.invoices.Find(ctx, id)
		if err != nil {
			return nil, err
		}
		if invoice == nil {
			return nil, ErrInvoiceNotFound
		}
		if invoice.State != InvoiceStatePosted {
			return nil, ErrInvoiceNotPosted
		}
		if invoice.AmountResidual.Round(4).IsZero() {
			continue
		}
		open = append(open, invoice)
	}
	return open, nil
}

func (s PaymentService) crossCurrency(open []*Invoice, paymentCurrency *string) bool {
	if s.orgs == nil || s.rates == nil || paymentCurrency == nil {
		return false
	}
	for _, inv := range open {
		if !currencyEqual(inv.CurrencyCode, paymentCurrency) {
			return true
		}
	}
	return false
}

func (s PaymentService) baseCurrency(ctx context.Context, organizationID uint64) (string, error) {
	if s.orgs == nil {
		return "", ErrOrganizationNotFound
	}
	org, err := s.orgs.Find(ctx, organizationID)
	if err != nil {
		return "", err
	}
	if org == nil {
		return "", ErrOrganizationNotFound
	}
	if org.BaseCurrency == "" {
		return "", ErrBaseCurrencyMissing
	}
	return org.BaseCurrency, nil
}

func derefCurrency(code *string, fallback string) string {
	if code == nil || *code == "" {
		return fallback
	}
	return *code
}

type crossAllocation struct {
	open         []*Invoice
	allocations  []*PaymentAllocation
	historicals  []amount.Amount
	settlements  []amount.Amount
	payableBase  amount.Amount
	receivableHt amount.Amount
	advanceBase  amount.Amount
}

func (s PaymentService) allocateCrossCurrency(ctx context.Context, organizationID uint64, base string, invoiceIDs []uint64, payAmount float64, payCurrency string, payDate time.Time, tolerance float64, allowAdvance bool) (*crossAllocation, error) {
	conv := amount.NewConverter(base, s.rates)
	payableBase, err := conv.Convert(ctx, amount.FromFloat64(payAmount).Round(4), payCurrency, base, organizationID, amount.RateSpot, payDate)
	if err != nil {
		return nil, err
	}
	payableBase = payableBase.Round(4)
	type need struct {
		invoice        *Invoice
		invCurrency    string
		residual       amount.Amount
		settlement     amount.Amount
		historicalFull amount.Amount
	}
	needs := make([]need, 0, len(invoiceIDs))
	totalSettle := amount.Zero()
	for _, id := range invoiceIDs {
		invoice, err := s.invoices.Find(ctx, id)
		if err != nil {
			return nil, err
		}
		if invoice == nil {
			return nil, ErrInvoiceNotFound
		}
		if invoice.State != InvoiceStatePosted {
			return nil, ErrInvoiceNotPosted
		}
		residual := invoice.AmountResidual.Round(4)
		if residual.IsZero() {
			continue
		}
		invCurrency := derefCurrency(invoice.CurrencyCode, base)
		settlement, err := conv.Convert(ctx, residual, invCurrency, base, organizationID, amount.RateSpot, payDate)
		if err != nil {
			return nil, err
		}
		settlement = settlement.Round(4)
		invoiceDate := payDate
		if invoice.InvoiceDate != nil && !invoice.InvoiceDate.IsZero() {
			invoiceDate = *invoice.InvoiceDate
		}
		historicalFull, err := conv.Convert(ctx, residual, invCurrency, base, organizationID, amount.RateSpot, invoiceDate)
		if err != nil {
			return nil, err
		}
		historicalFull = historicalFull.Round(4)
		needs = append(needs, need{invoice: invoice, invCurrency: invCurrency, residual: residual, settlement: settlement, historicalFull: historicalFull})
		totalSettle = totalSettle.Add(settlement)
	}
	over := payableBase.Sub(totalSettle).Round(4)
	if over.IsPositive() {
		tolBase, err := conv.Convert(ctx, amount.FromFloat64(tolerance).Abs(), payCurrency, base, organizationID, amount.RateSpot, payDate)
		if err != nil {
			return nil, err
		}
		if tolBase.IsZero() {
			tolBase = amount.FromFloat64(0.01)
		}
		if over.GreaterThan(tolBase.Round(4)) && !allowAdvance {
			return nil, ErrOverAllocation
		}
	}
	out := &crossAllocation{payableBase: payableBase}
	remaining := payableBase
	for _, n := range needs {
		allocSettle := n.settlement
		if remaining.LessThan(allocSettle) {
			allocSettle = remaining
		}
		allocForeign, err := conv.Convert(ctx, allocSettle, base, n.invCurrency, organizationID, amount.RateSpot, payDate)
		if err != nil {
			return nil, err
		}
		allocForeign = allocForeign.Round(4)
		if allocForeign.GreaterThan(n.residual) {
			allocForeign = n.residual
			allocSettle, err = conv.Convert(ctx, allocForeign, n.invCurrency, base, organizationID, amount.RateSpot, payDate)
			if err != nil {
				return nil, err
			}
			allocSettle = allocSettle.Round(4)
		}
		invoiceDate := payDate
		if n.invoice.InvoiceDate != nil && !n.invoice.InvoiceDate.IsZero() {
			invoiceDate = *n.invoice.InvoiceDate
		}
		historical, err := conv.Convert(ctx, allocForeign, n.invCurrency, base, organizationID, amount.RateSpot, invoiceDate)
		if err != nil {
			return nil, err
		}
		historical = historical.Round(4)
		n.invoice.AmountResidual = n.residual.Sub(allocForeign).Round(4)
		if n.invoice.AmountResidual.IsZero() {
			n.invoice.PaymentState = PaymentStatePaid
		} else {
			n.invoice.PaymentState = PaymentStatePartial
		}
		out.open = append(out.open, n.invoice)
		out.allocations = append(out.allocations, &PaymentAllocation{InvoiceID: n.invoice.ID, Amount: allocForeign.Float64()})
		out.historicals = append(out.historicals, historical)
		out.settlements = append(out.settlements, allocSettle)
		out.receivableHt = out.receivableHt.Add(historical)
		remaining = remaining.Sub(allocSettle).Round(4)
		if remaining.IsZero() {
			break
		}
	}
	if remaining.IsPositive() {
		out.advanceBase = remaining.Round(4)
	}
	return out, nil
}

func (s PaymentService) createCrossCurrency(ctx context.Context, tx *gorm.DB, request CreatePaymentRequest, debitAccount, creditAccount uint64, date time.Time, inbound bool) (*Payment, error) {
	if s.orgs == nil || s.rates == nil || s.fx == nil {
		return nil, ErrCurrencyMismatch
	}
	if request.CurrencyCode == nil || *request.CurrencyCode == "" {
		return nil, ErrCurrencyMismatch
	}
	if (request.DiscountAmount != 0 || request.WithholdingAmount != 0) && request.WriteOffAccountID == nil {
		return nil, ErrReconcileWriteOff
	}
	base, err := s.baseCurrency(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}
	payCurrency := *request.CurrencyCode
	conv := amount.NewConverter(base, s.rates)
	effectiveForeign := request.Amount + request.DiscountAmount + request.WithholdingAmount
	cross, err := s.allocateCrossCurrency(ctx, request.OrganizationID, base, request.InvoiceIDs, effectiveForeign, payCurrency, date, request.Tolerance, request.AllowAdvance)
	if err != nil {
		return nil, err
	}
	if len(cross.open) == 0 {
		return nil, ErrInvoiceNotFound
	}
	if !cross.advanceBase.IsZero() && !request.AllowAdvance {
		return nil, ErrOverAllocation
	}
	discountBase := amount.Zero()
	withholdingBase := amount.Zero()
	if request.DiscountAmount != 0 {
		discountBase, err = conv.Convert(ctx, amount.FromFloat64(request.DiscountAmount).Round(4), payCurrency, base, request.OrganizationID, amount.RateSpot, date)
		if err != nil {
			return nil, err
		}
		discountBase = discountBase.Round(4)
	}
	if request.WithholdingAmount != 0 {
		withholdingBase, err = conv.Convert(ctx, amount.FromFloat64(request.WithholdingAmount).Round(4), payCurrency, base, request.OrganizationID, amount.RateSpot, date)
		if err != nil {
			return nil, err
		}
		withholdingBase = withholdingBase.Round(4)
	}
	cashBase, err := conv.Convert(ctx, amount.FromFloat64(request.Amount).Round(4), payCurrency, base, request.OrganizationID, amount.RateSpot, date)
	if err != nil {
		return nil, err
	}
	cashBase = cashBase.Round(4)
	name, err := s.sequences.Next(ctx, request.OrganizationID, SequencePaymentCode)
	if err != nil {
		return nil, err
	}
	paymentType := PaymentTypeInbound
	description := "Customer payment"
	bankName := "Bank"
	counterName := "Accounts Receivable"
	if !inbound {
		paymentType = PaymentTypeOutbound
		description = "Supplier payment"
		counterName = "Accounts Payable"
	}
	payment := &Payment{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           helper.Ptr(name),
		ContactID:      request.ContactID,
		Type:           paymentType,
		JournalID:      helper.Ptr(request.JournalID),
		Amount:         request.Amount,
		CurrencyCode:   request.CurrencyCode,
		Date:           date,
		Reference:      helper.Ptr(request.Reference),
		State:          PaymentStatePosted,
	}
	var postLines []PostingLine
	if inbound {
		postLines = []PostingLine{
			{AccountID: debitAccount, Name: bankName, Debit: cashBase, CurrencyCode: request.CurrencyCode, AmountCurrency: request.Amount},
			{AccountID: creditAccount, Name: counterName, Credit: cross.receivableHt.Add(discountBase).Add(withholdingBase).Round(4), CurrencyCode: request.CurrencyCode, AmountCurrency: effectiveForeign},
		}
	} else {
		postLines = []PostingLine{
			{AccountID: debitAccount, Name: counterName, Debit: cross.receivableHt.Round(4), CurrencyCode: request.CurrencyCode, AmountCurrency: effectiveForeign},
			{AccountID: creditAccount, Name: bankName, Credit: cashBase, CurrencyCode: request.CurrencyCode, AmountCurrency: request.Amount},
		}
	}
	if request.DiscountAmount != 0 && request.WriteOffAccountID != nil {
		if inbound {
			postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Cash Discount", Debit: discountBase})
		} else {
			postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Cash Discount", Credit: discountBase})
		}
	}
	if request.WithholdingAmount != 0 && request.WriteOffAccountID != nil {
		if inbound {
			postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Withholding", Debit: withholdingBase})
		} else {
			postLines = append(postLines, PostingLine{AccountID: *request.WriteOffAccountID, Name: "Withholding", Credit: withholdingBase})
		}
	}
	fxLines, err := s.realizedFxLinesCross(ctx, request.OrganizationID, cross, inbound)
	if err != nil {
		return nil, err
	}
	postLines = append(postLines, fxLines...)
	if !cross.advanceBase.IsZero() {
		advanceAccount, err := s.advanceAccountID(ctx, request.OrganizationID, inbound)
		if err != nil {
			return nil, err
		}
		if inbound {
			postLines = append(postLines, PostingLine{AccountID: advanceAccount, Name: "Customer Advance", Credit: cross.advanceBase})
		} else {
			postLines = append(postLines, PostingLine{AccountID: advanceAccount, Name: "Supplier Advance", Debit: cross.advanceBase})
		}
	}
	return s.runPayment(ctx, tx, txPayment{payment: payment, date: date, name: name, journalID: request.JournalID, organizationID: request.OrganizationID, description: description, open: cross.open, allocations: cross.allocations, postLines: postLines})
}

func (s PaymentService) advanceAccountID(ctx context.Context, organizationID uint64, inbound bool) (uint64, error) {
	if s.advances != nil {
		if inbound {
			return s.advances.CustomerAdvanceAccountID(ctx, organizationID)
		}
		return s.advances.SupplierAdvanceAccountID(ctx, organizationID)
	}
	if inbound {
		return s.receivableAccount(ctx, organizationID)
	}
	return s.payableAccount(ctx, organizationID)
}

func (s PaymentService) realizedFxLinesCross(ctx context.Context, organizationID uint64, cross *crossAllocation, inbound bool) ([]PostingLine, error) {
	if s.fx == nil {
		return nil, nil
	}
	var lines []PostingLine
	for i := range cross.open {
		diff := cross.settlements[i].Sub(cross.historicals[i]).Round(4)
		if diff.IsZero() {
			continue
		}
		if !inbound {
			diff = diff.Neg()
		}
		if diff.IsPositive() {
			gainID, err := s.fx.FxGainAccountID(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			if gainID == 0 {
				continue
			}
			lines = append(lines, PostingLine{AccountID: gainID, Name: "FX Gain", Credit: diff})
		} else {
			lossID, err := s.fx.FxLossAccountID(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			if lossID == 0 {
				continue
			}
			lines = append(lines, PostingLine{AccountID: lossID, Name: "FX Loss", Debit: diff.Abs()})
		}
	}
	if len(lines) == 0 {
		return nil, nil
	}
	merged := map[uint64]PostingLine{}
	for _, line := range lines {
		if existing, ok := merged[line.AccountID]; ok {
			existing.Debit = existing.Debit.Add(line.Debit)
			existing.Credit = existing.Credit.Add(line.Credit)
			merged[line.AccountID] = existing
		} else {
			merged[line.AccountID] = line
		}
	}
	out := make([]PostingLine, 0, len(merged))
	for _, line := range merged {
		if line.Debit.IsZero() && line.Credit.IsZero() {
			continue
		}
		out = append(out, line)
	}
	return out, nil
}
