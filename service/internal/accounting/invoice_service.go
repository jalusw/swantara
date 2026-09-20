package accounting

import (
	"context"
	"strconv"
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

type AccountLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

type InvoiceLineRequest struct {
	ItemID         *uint64
	Description    string
	Qty            float64
	UnitID         *uint64
	UnitPrice      float64
	DiscountPct    float64
	TaxIDs         helper.Int64Array
	AccountID      uint64
	DimensionID    *uint64
	SaleLineID     *uint64
	PurchaseLineID *uint64
}

type CreateInvoiceRequest struct {
	OrganizationID  uint64
	JournalID       uint64
	ContactID       uint64
	Date            time.Time
	DueDate         *time.Time
	Reference       string
	DeferredAccount *uint64
	CurrencyCode    *string
	Draft           bool
	PaymentTermID   *uint64
	TaxRuleID       *uint64
	Lines           []InvoiceLineRequest
}

type CreateSupplierBillRequest struct {
	OrganizationID uint64
	JournalID      uint64
	ContactID      uint64
	Date           time.Time
	DueDate        *time.Time
	Reference      string
	CurrencyCode   *string
	Draft          bool
	PaymentTermID  *uint64
	TaxRuleID      *uint64
	Lines          []InvoiceLineRequest
}

type CreateCreditNoteRequest struct {
	OrganizationID    uint64
	OriginalInvoiceID uint64
	JournalID         uint64
	Date              time.Time
	DueDate           *time.Time
	Reference         string
}

type ItemResolver interface {
	FindTemplate(ctx context.Context, itemID uint64) (*struct {
		Type       string
		CategoryID *uint64
	}, error)
}

type InvoiceService struct {
	invoices           InvoiceDAO
	lines              InvoiceLineDAO
	taxes              InvoiceTaxDAO
	poster             Poster
	accounts           AccountLookup
	taxRef             dao.CRUD[reference.Tax]
	sequences          sequence.Service
	tx                 db.Transactioner
	now                func() time.Time
	products           ItemResolver
	categories         dao.CRUD[reference.ItemCategory]
	orgs               OrganizationLookup
	rates              CurrencyRateResolver
	terms              PaymentTermSplitter
	positions          TaxRuleMapper
	installments       InvoiceInstallmentDAO
	creditApplications InvoiceCreditApplicationDAO
	contraSettlements  InvoiceContraSettlementDAO
	downPayments       DownPaymentLinkDAO
	creditLimits       ContactCreditLimiter
}

func NewInvoiceService(
	invoices InvoiceDAO,
	lines InvoiceLineDAO,
	taxes InvoiceTaxDAO,
	poster Poster,
	accounts AccountLookup,
	taxRef dao.CRUD[reference.Tax],
	sequences sequence.Service,
	tx db.Transactioner,
) InvoiceService {
	return InvoiceService{
		invoices:  invoices,
		lines:     lines,
		taxes:     taxes,
		poster:    poster,
		accounts:  accounts,
		taxRef:    taxRef,
		sequences: sequences,
		tx:        tx,
		now:       time.Now,
	}
}

func (s InvoiceService) WithProducts(products ItemResolver, categories dao.CRUD[reference.ItemCategory]) InvoiceService {
	s.products = products
	s.categories = categories
	return s
}

func (s InvoiceService) WithCurrencyResolvers(orgs OrganizationLookup, rates CurrencyRateResolver) InvoiceService {
	s.orgs = orgs
	s.rates = rates
	return s
}

func (s InvoiceService) WithPaymentTerms(terms PaymentTermSplitter) InvoiceService {
	s.terms = terms
	return s
}

func (s InvoiceService) WithCreditLimits(limits ContactCreditLimiter) InvoiceService {
	s.creditLimits = limits
	return s
}

func (s InvoiceService) checkCreditLimit(ctx context.Context, organizationID, contactID uint64, total amount.Amount) error {
	if s.creditLimits == nil {
		return nil
	}
	limit, err := s.creditLimits.CreditLimit(ctx, contactID)
	if err != nil {
		return err
	}
	if limit == nil {
		return nil
	}
	open, err := s.invoices.ListOpenByContact(ctx, contactID)
	if err != nil {
		return err
	}
	exposure := total.Round(4)
	for _, invoice := range open {
		if invoice.OrganizationID == nil || *invoice.OrganizationID != organizationID {
			continue
		}
		exposure = exposure.Add(invoice.AmountResidual.Round(4))
	}
	if exposure.GreaterThan(amount.FromFloat64(*limit).Round(4)) {
		return ErrCreditLimitExceeded
	}
	return nil
}

func (s InvoiceService) WithTaxRules(positions TaxRuleMapper) InvoiceService {
	s.positions = positions
	return s
}

func (s InvoiceService) WithInstallments(installments InvoiceInstallmentDAO) InvoiceService {
	s.installments = installments
	return s
}

func (s InvoiceService) ListInstallments(ctx context.Context, invoiceID uint64) ([]*InvoiceInstallment, error) {
	if s.installments == nil {
		return nil, nil
	}
	return s.installments.ListByInvoice(ctx, invoiceID)
}

func (s InvoiceService) IsConfigured() bool {
	return s.invoices != nil
}

func (s InvoiceService) List(ctx context.Context, q *query.Query) (*query.Page[Invoice], error) {
	return s.invoices.List(ctx, q)
}

func (s InvoiceService) Get(ctx context.Context, id uint64) (*Invoice, error) {
	return s.invoices.Find(ctx, id)
}

func (s InvoiceService) ListLines(ctx context.Context, invoiceID uint64) ([]*InvoiceLine, error) {
	return s.lines.ListByInvoice(ctx, invoiceID)
}

func (s InvoiceService) ListTaxes(ctx context.Context, invoiceID uint64) ([]*InvoiceTax, error) {
	return s.taxes.ListByInvoice(ctx, invoiceID)
}

func (s InvoiceService) Create(ctx context.Context, request CreateInvoiceRequest) (*Invoice, error) {
	var invoice *Invoice
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		invoice, err = s.CreateTx(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s InvoiceService) CreateTx(ctx context.Context, tx *gorm.DB, request CreateInvoiceRequest) (*Invoice, error) {
	lines, err := s.applyTaxRule(ctx, request.OrganizationID, request.TaxRuleID, request.Lines)
	if err != nil {
		return nil, err
	}
	computed, err := s.computeInvoice(ctx, request.OrganizationID, lines, reference.TaxScopeSale)
	if err != nil {
		return nil, err
	}

	total := computed.untaxed.Add(computed.tax).Round(4)
	untaxed := computed.untaxed.Round(4)
	tax := computed.tax.Round(4)

	if err := s.checkCreditLimit(ctx, request.OrganizationID, request.ContactID, total); err != nil {
		return nil, err
	}

	name, err := s.sequences.NextTx(ctx, tx, request.OrganizationID, SequenceInvoiceCode)
	if err != nil {
		return nil, err
	}
	invoiceDate := request.Date
	if invoiceDate.IsZero() {
		invoiceDate = s.now().UTC()
	}
	splits, err := s.paymentSplits(ctx, request.PaymentTermID, total, invoiceDate)
	if err != nil {
		return nil, err
	}
	dueDate := resolveSplitDueDate(request.DueDate, splits)

	invoice := &Invoice{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Type:           InvoiceTypeCustomerInvoice,
		ContactID:      request.ContactID,
		Name:           helper.Ptr(name),
		Reference:      helper.Ptr(request.Reference),
		InvoiceDate:    &invoiceDate,
		DueDate:        dueDate,
		CurrencyCode:   request.CurrencyCode,
		JournalID:      helper.Ptr(request.JournalID),
		PaymentTermID:  request.PaymentTermID,
		TaxRuleID:      request.TaxRuleID,
		State:          InvoiceStatePosted,
		PaymentState:   PaymentStateNotPaid,
		AmountUntaxed:  untaxed,
		AmountTax:      tax,
		AmountTotal:    total,
		AmountResidual: total,
	}
	if request.TaxRuleID != nil {
		invoice.TaxRule = helper.Ptr(formatTaxRule(*request.TaxRuleID))
	}
	if request.Draft {
		invoice.State = InvoiceStateDraft
		created, err := s.invoices.CreateWithLinesTx(ctx, tx, invoice, computed.lines, s.invoiceTaxes(computed))
		if err != nil {
			return nil, err
		}
		if err := s.persistInstallments(ctx, tx, created, splits); err != nil {
			return nil, err
		}
		return created, nil
	}

	postLines, err := s.customerPostLines(ctx, request.OrganizationID, request.ContactID, request.CurrencyCode, dueDate, request.DeferredAccount, invoiceDate, computed)
	if err != nil {
		return nil, err
	}

	entry, err := s.poster.PostTx(ctx, tx, PostRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Date:           invoiceDate,
		Ref:            name,
		OriginType:     InvoiceTypeCustomerInvoice,
		Description:    "Customer invoice",
		CurrencyCode:   request.CurrencyCode,
		Lines:          postLines,
	})
	if err != nil {
		return nil, err
	}
	invoice.EntryID = helper.Ptr(entry.ID)

	created, err := s.invoices.CreateWithLinesTx(ctx, tx, invoice, computed.lines, s.invoiceTaxes(computed))
	if err != nil {
		return nil, err
	}
	if err := s.persistInstallments(ctx, tx, created, splits); err != nil {
		return nil, err
	}
	return created, nil
}

func (s InvoiceService) customerPostLines(ctx context.Context, organizationID, contactID uint64, currencyCode *string, dueDate *time.Time, deferredAccount *uint64, invoiceDate time.Time, computed *computedInvoice) ([]PostingLine, error) {
	receivable, err := s.receivableAccount(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	total := computed.untaxed.Add(computed.tax).Round(4)
	untaxed := computed.untaxed.Round(4)

	functionalTotal, err := s.functional(ctx, organizationID, total, currencyCode, invoiceDate)
	if err != nil {
		return nil, err
	}
	functionalUntaxed, err := s.functional(ctx, organizationID, untaxed, currencyCode, invoiceDate)
	if err != nil {
		return nil, err
	}
	postLines := []PostingLine{
		{AccountID: receivable, Name: "Accounts Receivable", Debit: functionalTotal, CurrencyCode: currencyCode, AmountCurrency: total.Float64(), ContactID: helper.Ptr(contactID), DueDate: dueDate},
	}
	if deferredAccount != nil {
		postLines = append(postLines, PostingLine{AccountID: *deferredAccount, Name: "Deferred Revenue", Credit: functionalUntaxed, CurrencyCode: currencyCode, AmountCurrency: untaxed.Float64()})
	} else {
		revenueByAccount := map[uint64]amount.Amount{}
		for _, line := range computed.lines {
			revenueByAccount[*line.AccountID] = revenueByAccount[*line.AccountID].Add(line.PriceSubtotal)
		}
		for accountID, revenue := range revenueByAccount {
			functionalRevenue, err := s.functional(ctx, organizationID, revenue, currencyCode, invoiceDate)
			if err != nil {
				return nil, err
			}
			postLines = append(postLines, PostingLine{AccountID: accountID, Name: "Revenue", Credit: functionalRevenue, CurrencyCode: currencyCode, AmountCurrency: revenue.Float64()})
		}
	}
	for accountID, taxAmount := range computed.taxByAccount {
		functionalTax, err := s.functional(ctx, organizationID, taxAmount, currencyCode, invoiceDate)
		if err != nil {
			return nil, err
		}
		postLines = append(postLines, PostingLine{AccountID: accountID, Name: "Output Tax", Credit: functionalTax, CurrencyCode: currencyCode, AmountCurrency: taxAmount.Float64()})
	}
	return postLines, nil
}

func (s InvoiceService) CreateSupplierBill(ctx context.Context, request CreateSupplierBillRequest) (*Invoice, error) {
	lines, err := s.applyTaxRule(ctx, request.OrganizationID, request.TaxRuleID, request.Lines)
	if err != nil {
		return nil, err
	}
	computed, err := s.computeInvoice(ctx, request.OrganizationID, lines, reference.TaxScopePurchase)
	if err != nil {
		return nil, err
	}

	payable, err := s.payableAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}

	total := computed.untaxed.Add(computed.tax).Round(4)
	untaxed := computed.untaxed.Round(4)
	tax := computed.tax.Round(4)

	name, err := s.sequences.Next(ctx, request.OrganizationID, SequenceInvoiceCode)
	if err != nil {
		return nil, err
	}
	invoiceDate := request.Date
	if invoiceDate.IsZero() {
		invoiceDate = s.now().UTC()
	}
	splits, err := s.paymentSplits(ctx, request.PaymentTermID, total, invoiceDate)
	if err != nil {
		return nil, err
	}
	dueDate := resolveSplitDueDate(request.DueDate, splits)

	invoice := &Invoice{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Type:           InvoiceTypeSupplierBill,
		ContactID:      request.ContactID,
		Name:           helper.Ptr(name),
		Reference:      helper.Ptr(request.Reference),
		InvoiceDate:    &invoiceDate,
		DueDate:        dueDate,
		CurrencyCode:   request.CurrencyCode,
		JournalID:      helper.Ptr(request.JournalID),
		PaymentTermID:  request.PaymentTermID,
		TaxRuleID:      request.TaxRuleID,
		State:          InvoiceStatePosted,
		PaymentState:   PaymentStateNotPaid,
		AmountUntaxed:  untaxed,
		AmountTax:      tax,
		AmountTotal:    total,
		AmountResidual: total,
	}
	if request.TaxRuleID != nil {
		invoice.TaxRule = helper.Ptr(formatTaxRule(*request.TaxRuleID))
	}
	if request.Draft {
		invoice.State = InvoiceStateDraft
	}

	var created *Invoice
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		if request.Draft {
			var err error
			created, err = s.invoices.CreateWithLinesTx(ctx, tx, invoice, computed.lines, s.invoiceTaxes(computed))
			if err != nil {
				return err
			}
			return s.persistInstallments(ctx, tx, created, splits)
		}
		postLines, err := s.supplierPostLines(ctx, request.OrganizationID, request.ContactID, request.CurrencyCode, dueDate, invoiceDate, payable, computed)
		if err != nil {
			return err
		}

		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      request.JournalID,
			Date:           invoiceDate,
			Ref:            name,
			OriginType:     InvoiceTypeSupplierBill,
			Description:    "Supplier bill",
			CurrencyCode:   request.CurrencyCode,
			Lines:          postLines,
		})
		if err != nil {
			return err
		}
		invoice.EntryID = helper.Ptr(entry.ID)

		created, err = s.invoices.CreateWithLinesTx(ctx, tx, invoice, computed.lines, s.invoiceTaxes(computed))
		if err != nil {
			return err
		}
		return s.persistInstallments(ctx, tx, created, splits)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s InvoiceService) supplierPostLines(ctx context.Context, organizationID, contactID uint64, currencyCode *string, dueDate *time.Time, invoiceDate time.Time, payable uint64, computed *computedInvoice) ([]PostingLine, error) {
	total := computed.untaxed.Add(computed.tax).Round(4)
	functionalTotal, err := s.functional(ctx, organizationID, total, currencyCode, invoiceDate)
	if err != nil {
		return nil, err
	}
	postLines := []PostingLine{
		{AccountID: payable, Name: "Accounts Payable", Credit: functionalTotal, CurrencyCode: currencyCode, AmountCurrency: total.Float64(), ContactID: helper.Ptr(contactID), DueDate: dueDate},
	}
	expenseByAccount := map[uint64]amount.Amount{}
	for _, line := range computed.lines {
		accID := s.resolveVendorAccount(ctx, line)
		expenseByAccount[accID] = expenseByAccount[accID].Add(line.PriceSubtotal)
	}
	for accountID, expense := range expenseByAccount {
		functionalExpense, err := s.functional(ctx, organizationID, expense, currencyCode, invoiceDate)
		if err != nil {
			return nil, err
		}
		postLines = append(postLines, PostingLine{AccountID: accountID, Name: "Expense", Debit: functionalExpense, CurrencyCode: currencyCode, AmountCurrency: expense.Float64()})
	}
	for accountID, taxAmount := range computed.taxByAccount {
		functionalTax, err := s.functional(ctx, organizationID, taxAmount, currencyCode, invoiceDate)
		if err != nil {
			return nil, err
		}
		postLines = append(postLines, PostingLine{AccountID: accountID, Name: "Input Tax", Debit: functionalTax, CurrencyCode: currencyCode, AmountCurrency: taxAmount.Float64()})
	}
	return postLines, nil
}

func (s InvoiceService) CreateCreditNote(ctx context.Context, request CreateCreditNoteRequest) (*Invoice, error) {
	var created *Invoice
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = s.CreateCreditNoteTx(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s InvoiceService) CreateCreditNoteTx(ctx context.Context, tx *gorm.DB, request CreateCreditNoteRequest) (*Invoice, error) {
	return s.createCreditNoteTx(ctx, tx, request, InvoiceTypeCustomerInvoice, InvoiceTypeCustomerCredit)
}

func (s InvoiceService) Cancel(ctx context.Context, invoiceID, organizationID uint64) (*Invoice, error) {
	invoice, err := s.invoices.Find(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil || *invoice.OrganizationID != organizationID {
		return nil, ErrInvoiceNotFound
	}
	if !invoice.AmountResidual.Equal(invoice.AmountTotal) {
		return nil, ErrInvoicePaid
	}
	if invoice.State == InvoiceStateDraft {
		invoice.State = InvoiceStateCancelled
		return s.invoices.Update(ctx, invoice)
	}
	if invoice.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}
	if invoice.EntryID == nil {
		return nil, ErrEntryNotFound
	}
	if invoice.JournalID == nil {
		return nil, ErrNoReceivableAccount
	}
	if _, err := s.poster.Reverse(ctx, ReverseRequest{
		OrganizationID: organizationID,
		JournalID:      *invoice.JournalID,
		Date:           s.now().UTC(),
		Ref:            *invoice.Name + "/CANCEL",
		Description:    "Cancel invoice",
		EntryID:        *invoice.EntryID,
	}); err != nil {
		return nil, err
	}
	invoice.State = InvoiceStateCancelled
	return s.invoices.Update(ctx, invoice)
}

func (s InvoiceService) CreateVendorCreditNote(ctx context.Context, request CreateCreditNoteRequest) (*Invoice, error) {
	var created *Invoice
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = s.CreateVendorCreditNoteTx(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s InvoiceService) CreateVendorCreditNoteTx(ctx context.Context, tx *gorm.DB, request CreateCreditNoteRequest) (*Invoice, error) {
	return s.createCreditNoteTx(ctx, tx, request, InvoiceTypeSupplierBill, InvoiceTypeSupplierCredit)
}

func (s InvoiceService) createCreditNoteTx(ctx context.Context, tx *gorm.DB, request CreateCreditNoteRequest, originType, creditType string) (*Invoice, error) {
	original, err := s.invoices.Find(ctx, request.OriginalInvoiceID)
	if err != nil {
		return nil, err
	}
	if original == nil || original.OrganizationID == nil || *original.OrganizationID != request.OrganizationID {
		return nil, ErrInvoiceNotFound
	}
	if original.Type != originType || original.State != InvoiceStatePosted {
		return nil, ErrInvoiceNotPosted
	}

	existing, err := s.invoices.FindByOrigin(ctx, original.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrInvoiceReversed
	}

	originalLines, err := s.lines.ListByInvoice(ctx, original.ID)
	if err != nil {
		return nil, err
	}
	originalTaxes, err := s.taxes.ListByInvoice(ctx, original.ID)
	if err != nil {
		return nil, err
	}
	if len(originalLines) == 0 {
		return nil, ErrInvoiceNoLines
	}

	name, err := s.sequences.Next(ctx, request.OrganizationID, SequenceInvoiceCode)
	if err != nil {
		return nil, err
	}
	invoiceDate := request.Date
	if invoiceDate.IsZero() {
		invoiceDate = s.now().UTC()
	}

	creditNote := &Invoice{
		OrganizationID:  helper.Ptr(request.OrganizationID),
		Type:            creditType,
		ContactID:       original.ContactID,
		Name:            helper.Ptr(name),
		Reference:       helper.Ptr(request.Reference),
		InvoiceDate:     &invoiceDate,
		DueDate:         request.DueDate,
		JournalID:       helper.Ptr(request.JournalID),
		State:           InvoiceStatePosted,
		PaymentState:    PaymentStateNotPaid,
		AmountUntaxed:   original.AmountUntaxed,
		AmountTax:       original.AmountTax,
		AmountTotal:     original.AmountTotal,
		AmountResidual:  original.AmountTotal,
		OriginInvoiceID: helper.Ptr(original.ID),
	}

	lines := make([]*InvoiceLine, len(originalLines))
	for i, line := range originalLines {
		lines[i] = &InvoiceLine{
			Sequence:       line.Sequence,
			ItemID:         line.ItemID,
			Description:    line.Description,
			Qty:            line.Qty,
			UnitID:         line.UnitID,
			UnitPrice:      line.UnitPrice,
			DiscountPct:    line.DiscountPct,
			TaxIDs:         line.TaxIDs,
			AccountID:      line.AccountID,
			DimensionID:    line.DimensionID,
			PriceSubtotal:  line.PriceSubtotal,
			SaleLineID:     line.SaleLineID,
			PurchaseLineID: line.PurchaseLineID,
		}
	}
	taxes := make([]*InvoiceTax, len(originalTaxes))
	for i, tax := range originalTaxes {
		taxes[i] = &InvoiceTax{
			TaxID:      tax.TaxID,
			BaseAmount: tax.BaseAmount,
			Amount:     tax.Amount,
			AccountID:  tax.AccountID,
		}
	}

	entry, err := s.poster.ReverseTx(ctx, tx, ReverseRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Date:           invoiceDate,
		Ref:            name,
		Description:    "Credit note",
		EntryID:        *original.EntryID,
	})
	if err != nil {
		return nil, err
	}
	creditNote.EntryID = helper.Ptr(entry.ID)

	return s.invoices.CreateWithLinesTx(ctx, tx, creditNote, lines, taxes)
}

type computedInvoice struct {
	untaxed          amount.Amount
	tax              amount.Amount
	lines            []*InvoiceLine
	taxByAccount     map[uint64]amount.Amount
	taxBaseByAccount map[uint64]amount.Amount
	taxIDByAccount   map[uint64]uint64
}

func (s InvoiceService) computeInvoice(ctx context.Context, organizationID uint64, requestLines []InvoiceLineRequest, scope string) (*computedInvoice, error) {
	if len(requestLines) == 0 {
		return nil, ErrInvoiceNoLines
	}
	for _, line := range requestLines {
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrInvoiceLineQty
		}
	}

	untaxed := amount.Zero()
	tax := amount.Zero()
	taxByAccount := map[uint64]amount.Amount{}
	taxBaseByAccount := map[uint64]amount.Amount{}
	taxIDByAccount := map[uint64]uint64{}
	lines := make([]*InvoiceLine, 0, len(requestLines))
	for _, line := range requestLines {
		qty := amount.FromFloat64(line.Qty)
		discount := amount.FromFloat64(1 - line.DiscountPct/100)
		subtotal := qty.Mul(amount.FromFloat64(line.UnitPrice)).Mul(discount).Round(4)
		untaxed = untaxed.Add(subtotal)

		for _, taxID := range line.TaxIDs {
			taxEntity, err := s.taxRef.Find(ctx, uint64(taxID))
			if err != nil {
				return nil, err
			}
			if taxEntity == nil {
				return nil, ErrInvoiceTaxInvalid
			}
			if taxEntity.Scope != scope && taxEntity.Scope != reference.TaxScopeNone {
				return nil, ErrInvoiceTaxInvalid
			}
			taxAmount, err := reference.TaxLineAmount(subtotal, qty, taxEntity)
			if err != nil {
				return nil, err
			}
			if taxEntity.TaxAccountID == nil {
				return nil, ErrInvoiceTaxInvalid
			}
			taxByAccount[*taxEntity.TaxAccountID] = taxByAccount[*taxEntity.TaxAccountID].Add(taxAmount)
			taxBaseByAccount[*taxEntity.TaxAccountID] = taxBaseByAccount[*taxEntity.TaxAccountID].Add(subtotal)
			taxIDByAccount[*taxEntity.TaxAccountID] = uint64(taxID)
		}

		lines = append(lines, &InvoiceLine{
			Sequence:       len(lines)*10 + 10,
			ItemID:         line.ItemID,
			Description:    helper.Ptr(line.Description),
			Qty:            line.Qty,
			UnitID:         line.UnitID,
			UnitPrice:      amount.FromFloat64(line.UnitPrice),
			DiscountPct:    line.DiscountPct,
			TaxIDs:         line.TaxIDs,
			AccountID:      helper.Ptr(line.AccountID),
			DimensionID:    line.DimensionID,
			PriceSubtotal:  subtotal,
			SaleLineID:     line.SaleLineID,
			PurchaseLineID: line.PurchaseLineID,
		})
	}

	for _, taxAmount := range taxByAccount {
		tax = tax.Add(taxAmount)
	}
	return &computedInvoice{
		untaxed:          untaxed,
		tax:              tax,
		lines:            lines,
		taxByAccount:     taxByAccount,
		taxBaseByAccount: taxBaseByAccount,
		taxIDByAccount:   taxIDByAccount,
	}, nil
}

func (s InvoiceService) invoiceTaxes(computed *computedInvoice) []*InvoiceTax {
	taxes := make([]*InvoiceTax, 0, len(computed.taxByAccount))
	for accountID, taxAmount := range computed.taxByAccount {
		taxes = append(taxes, &InvoiceTax{
			TaxID:      helper.Ptr(computed.taxIDByAccount[accountID]),
			BaseAmount: computed.taxBaseByAccount[accountID].Round(4),
			Amount:     taxAmount.Round(4),
			AccountID:  helper.Ptr(accountID),
		})
	}
	return taxes
}

func (s InvoiceService) applyTaxRule(ctx context.Context, organizationID uint64, positionID *uint64, lines []InvoiceLineRequest) ([]InvoiceLineRequest, error) {
	if positionID == nil || s.positions == nil {
		return lines, nil
	}
	position, err := s.positions.Find(ctx, *positionID)
	if err != nil {
		return nil, err
	}
	if position == nil || position.OrganizationID == nil || *position.OrganizationID != organizationID {
		return nil, ErrTaxRuleNotFound
	}
	mapped := make([]InvoiceLineRequest, len(lines))
	for i, line := range lines {
		mapped[i] = line
		accountResult, err := s.positions.Resolve(ctx, *positionID, nil, line.AccountID)
		if err != nil {
			return nil, err
		}
		if accountResult != nil && accountResult.Account != nil {
			mapped[i].AccountID = *accountResult.Account
		}
		if len(line.TaxIDs) == 0 {
			continue
		}
		remapped := make(helper.Int64Array, 0, len(line.TaxIDs))
		for _, taxID := range line.TaxIDs {
			src := uint64(taxID)
			result, err := s.positions.Resolve(ctx, *positionID, &src, line.AccountID)
			if err != nil {
				return nil, err
			}
			if result == nil || result.TaxAccount == nil {
				continue
			}
			remapped = append(remapped, int64(*result.TaxAccount))
		}
		mapped[i].TaxIDs = remapped
	}
	return mapped, nil
}

func (s InvoiceService) paymentSplits(ctx context.Context, termID *uint64, total amount.Amount, invoiceDate time.Time) ([]PaymentTermSplit, error) {
	if termID == nil || s.terms == nil {
		return nil, nil
	}
	return s.terms.Splits(ctx, *termID, total, invoiceDate)
}

func resolveSplitDueDate(explicit *time.Time, splits []PaymentTermSplit) *time.Time {
	if explicit != nil {
		return explicit
	}
	if len(splits) == 0 {
		return nil
	}
	latest := splits[0].DueDate
	for _, split := range splits[1:] {
		if split.DueDate.After(latest) {
			latest = split.DueDate
		}
	}
	return &latest
}

func (s InvoiceService) persistInstallments(ctx context.Context, tx *gorm.DB, invoice *Invoice, splits []PaymentTermSplit) error {
	if s.installments == nil || invoice == nil || len(splits) == 0 {
		return nil
	}
	records := make([]*InvoiceInstallment, len(splits))
	for i, split := range splits {
		due := split.DueDate
		records[i] = &InvoiceInstallment{
			OrganizationID: invoice.OrganizationID,
			InvoiceID:      invoice.ID,
			Sequence:       split.Sequence,
			DueDate:        &due,
			Amount:         split.Amount.Round(4),
			State:          InstallmentStatePending,
		}
		if records[i].Sequence == 0 {
			records[i].Sequence = (i + 1) * 10
		}
	}
	return s.installments.CreateManyTx(ctx, tx, records)
}

func formatTaxRule(positionID uint64) string {
	return "tax_rule:" + strconv.FormatUint(positionID, 10)
}

func (s InvoiceService) resolveVendorAccount(ctx context.Context, line *InvoiceLine) uint64 {
	if line.AccountID == nil {
		return 0
	}
	if s.products == nil || s.categories == nil || line.ItemID == nil {
		return *line.AccountID
	}
	tpl, err := s.products.FindTemplate(ctx, *line.ItemID)
	if err != nil || tpl == nil {
		return *line.AccountID
	}
	if tpl.Type != "storable" && tpl.Type != "stockable" {
		return *line.AccountID
	}
	if tpl.CategoryID == nil {
		return *line.AccountID
	}
	cat, err := s.categories.Find(ctx, *tpl.CategoryID)
	if err != nil || cat == nil || cat.StockValuationAccountID == nil {
		return *line.AccountID
	}
	return *cat.StockValuationAccountID
}

func (s InvoiceService) functional(ctx context.Context, organizationID uint64, foreign amount.Amount, currencyCode *string, date time.Time) (amount.Amount, error) {
	if currencyCode == nil || *currencyCode == "" || s.orgs == nil || s.rates == nil {
		return foreign.Round(4), nil
	}
	org, err := s.orgs.Find(ctx, organizationID)
	if err != nil {
		return amount.Amount{}, err
	}
	if org == nil {
		return amount.Amount{}, ErrOrganizationNotFound
	}
	if org.BaseCurrency == "" {
		return amount.Amount{}, ErrBaseCurrencyMissing
	}
	if *currencyCode == org.BaseCurrency {
		return foreign.Round(4), nil
	}
	conv := amount.NewConverter(org.BaseCurrency, s.rates)
	converted, err := conv.Convert(ctx, foreign.Round(4), *currencyCode, org.BaseCurrency, organizationID, amount.RateSpot, date)
	if err != nil {
		return amount.Amount{}, err
	}
	return converted.Round(4), nil
}

func (s InvoiceService) payableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
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

func (s InvoiceService) receivableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
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
