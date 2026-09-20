package expense

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type ExpenseService struct {
	reports    ExpenseReportDAO
	lines      ExpenseLineDAO
	categories dao.CRUD[reference.ExpenseCategory]
	taxes      dao.CRUD[reference.Tax]
	poster     accounting.Poster
	configs    ExpenseConfigSource
	invoices   InvoiceEngine
	products   IncomeAccountResolver
	projects   ProjectLookup
	tx         db.Transactioner
	now        func() time.Time
}

func NewExpenseService(
	reports ExpenseReportDAO,
	lines ExpenseLineDAO,
	categories dao.CRUD[reference.ExpenseCategory],
	taxes dao.CRUD[reference.Tax],
	poster accounting.Poster,
	configs ExpenseConfigSource,
	tx db.Transactioner,
) ExpenseService {
	return ExpenseService{
		reports:    reports,
		lines:      lines,
		categories: categories,
		taxes:      taxes,
		poster:     poster,
		configs:    configs,
		tx:         tx,
		now:        time.Now,
	}
}

func (s ExpenseService) SetBilling(invoices InvoiceEngine, products IncomeAccountResolver, projects ProjectLookup) ExpenseService {
	s.invoices = invoices
	s.products = products
	s.projects = projects
	return s
}

type CreateExpenseLineRequest struct {
	EmployeeID          *uint64
	CategoryID          *uint64
	ItemID              *uint64
	Description         string
	ExpenseDate         time.Time
	Quantity            float64
	UnitPrice           float64
	TaxIDs              helper.Int64Array
	CurrencyCode        string
	DimensionID         *uint64
	ProjectID           *uint64
	Reimbursable        bool
	ReceiptAttachmentID *uint64
}

type CreateExpenseReportRequest struct {
	OrganizationID uint64
	Name           string
	EmployeeID     uint64
	PaymentMode    string
	Lines          []CreateExpenseLineRequest
}

func (s ExpenseService) Create(ctx context.Context, request CreateExpenseReportRequest) (*ExpenseReport, error) {
	if len(request.Lines) == 0 {
		return nil, ErrExpenseNoLines
	}
	if request.PaymentMode != ExpensePaymentOwnAccount && request.PaymentMode != ExpensePaymentOrgAccount {
		return nil, ErrExpenseInvalidLine
	}
	report := &ExpenseReport{
		OrganizationID: &request.OrganizationID,
		Name:           request.Name,
		EmployeeID:     request.EmployeeID,
		State:          ExpenseStateDraft,
		PaymentMode:    request.PaymentMode,
	}
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		created, err := s.reports.CreateTx(ctx, tx, report)
		if err != nil {
			return err
		}
		report = created
		total := 0.0
		for _, lineRequest := range request.Lines {
			line := &ExpenseLine{
				ReportID:            report.ID,
				EmployeeID:          lineRequest.EmployeeID,
				CategoryID:          lineRequest.CategoryID,
				ItemID:              lineRequest.ItemID,
				Description:         &lineRequest.Description,
				ExpenseDate:         &lineRequest.ExpenseDate,
				Quantity:            lineRequest.Quantity,
				UnitPrice:           lineRequest.UnitPrice,
				Amount:              lineRequest.Quantity * lineRequest.UnitPrice,
				TaxIDs:              lineRequest.TaxIDs,
				CurrencyCode:        &lineRequest.CurrencyCode,
				DimensionID:         lineRequest.DimensionID,
				ProjectID:           lineRequest.ProjectID,
				Reimbursable:        lineRequest.Reimbursable,
				ReceiptAttachmentID: lineRequest.ReceiptAttachmentID,
			}
			createdLine, err := s.lines.CreateTx(ctx, tx, line)
			if err != nil {
				return err
			}
			total += createdLine.Amount
		}
		report.TotalAmount = total
		updated, err := s.reports.UpdateTx(ctx, tx, report)
		if err != nil {
			return err
		}
		report = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s ExpenseService) Get(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	return s.findReport(ctx, organizationID, reportID)
}

func (s ExpenseService) List(ctx context.Context, organizationID uint64) ([]*ExpenseReport, error) {
	page, err := s.reports.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s ExpenseService) ListLines(ctx context.Context, organizationID, reportID uint64) ([]*ExpenseLine, error) {
	if _, err := s.findReport(ctx, organizationID, reportID); err != nil {
		return nil, err
	}
	return s.lines.ListByReport(ctx, reportID)
}

func (s ExpenseService) Submit(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStateDraft {
		return nil, ErrExpenseReportState
	}
	report.State = ExpenseStateSubmitted
	report.SubmittedAt = s.ptrTime(s.now())
	return s.reports.Update(ctx, report)
}

func (s ExpenseService) Approve(ctx context.Context, organizationID, reportID, approverID uint64) (*ExpenseReport, error) {
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStateSubmitted {
		return nil, ErrExpenseReportState
	}
	report.State = ExpenseStateApproved
	report.ApprovedBy = &approverID
	return s.reports.Update(ctx, report)
}

func (s ExpenseService) Refuse(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStateSubmitted {
		return nil, ErrExpenseReportState
	}
	report.State = ExpenseStateRefused
	return s.reports.Update(ctx, report)
}

func (s ExpenseService) Post(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStateApproved {
		return nil, ErrExpenseReportState
	}
	lines, err := s.lines.ListByReport(ctx, report.ID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrExpenseNoLines
	}
	journalID, err := s.configs.JournalID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if journalID == 0 {
		return nil, ErrExpenseConfig
	}
	payableAccountID, err := s.payableAccount(ctx, report)
	if err != nil {
		return nil, err
	}

	postLines, total, err := s.buildExpensePostingLines(ctx, organizationID, lines)
	if err != nil {
		return nil, err
	}
	postLines = append(postLines, accounting.PostingLine{AccountID: payableAccountID, Name: "Expense payable", Credit: amount.FromFloat64(total)})

	request := accounting.PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           s.now().UTC(),
		Ref:            report.Name,
		OriginType:     accounting.OriginTypeExpenseReport,
		OriginID:       report.ID,
		Description:    "Expense report " + report.Name,
		Lines:          postLines,
	}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		journal, err := s.poster.PostTx(ctx, tx, request)
		if err != nil {
			return err
		}
		report.EntryID = &journal.ID
		report.TotalAmount = total
		report.State = ExpenseStatePosted
		updated, err := s.reports.UpdateTx(ctx, tx, report)
		if err != nil {
			return err
		}
		report = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s ExpenseService) Reimburse(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStatePosted {
		return nil, ErrExpenseReportState
	}
	if report.PaymentMode != ExpensePaymentOwnAccount {
		return nil, ErrExpenseReportState
	}
	journalID, err := s.configs.JournalID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if journalID == 0 {
		return nil, ErrExpenseConfig
	}
	bankAccountID, err := s.configs.ReimbursementBankAccountID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if bankAccountID == 0 {
		return nil, ErrExpenseConfig
	}
	payableAccountID, err := s.payableAccount(ctx, report)
	if err != nil {
		return nil, err
	}

	request := accounting.PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           s.now().UTC(),
		Ref:            report.Name,
		OriginType:     accounting.OriginTypeExpenseReport,
		OriginID:       report.ID,
		Description:    "Expense reimbursement " + report.Name,
		Lines: []accounting.PostingLine{
			{AccountID: payableAccountID, Name: "Employee reimbursement", Debit: amount.FromFloat64(report.TotalAmount)},
			{AccountID: bankAccountID, Name: "Employee reimbursement", Credit: amount.FromFloat64(report.TotalAmount)},
		},
	}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		journal, err := s.poster.PostTx(ctx, tx, request)
		if err != nil {
			return err
		}
		report.ReimbursementEntryID = &journal.ID
		report.State = ExpenseStateReimbursed
		updated, err := s.reports.UpdateTx(ctx, tx, report)
		if err != nil {
			return err
		}
		report = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s ExpenseService) BillToInvoice(ctx context.Context, organizationID, reportID uint64) (*accounting.Invoice, error) {
	if s.invoices == nil || s.products == nil || s.projects == nil {
		return nil, ErrExpenseConfig
	}
	report, err := s.findReport(ctx, organizationID, reportID)
	if err != nil {
		return nil, err
	}
	if report.State != ExpenseStatePosted {
		return nil, ErrExpenseReportState
	}
	lines, err := s.lines.ListByReport(ctx, report.ID)
	if err != nil {
		return nil, err
	}
	var billable []*ExpenseLine
	for _, line := range lines {
		if !line.Reimbursable {
			billable = append(billable, line)
		}
	}
	if len(billable) == 0 {
		return nil, ErrExpenseNoBillable
	}
	var projectID uint64
	if billable[0].ProjectID == nil {
		return nil, ErrExpenseBillableProject
	}
	projectID = *billable[0].ProjectID
	for _, line := range billable {
		if line.ProjectID == nil || *line.ProjectID != projectID {
			return nil, ErrExpenseBillableProject
		}
	}
	project, err := s.projects.Search(ctx, "id", projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrExpenseProjectNotFound
	}
	journalID, err := s.configs.JournalID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if journalID == 0 {
		return nil, ErrExpenseConfig
	}

	invoiceLines := make([]accounting.InvoiceLineRequest, 0, len(billable))
	for _, line := range billable {
		if line.ItemID == nil {
			return nil, ErrExpenseNoIncomeAccount
		}
		incomeAccount, err := s.products.ResolveIncomeAccount(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		if incomeAccount == 0 {
			return nil, ErrExpenseNoIncomeAccount
		}
		description := ""
		if line.Description != nil {
			description = *line.Description
		}
		invoiceLines = append(invoiceLines, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: description,
			Qty:         line.Quantity,
			UnitPrice:   line.UnitPrice,
			TaxIDs:      line.TaxIDs,
			AccountID:   incomeAccount,
			DimensionID: line.DimensionID,
		})
	}
	return s.invoices.Create(ctx, accounting.CreateInvoiceRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		ContactID:      project.ContactID,
		Date:           s.now().UTC(),
		Reference:      report.Name,
		Lines:          invoiceLines,
	})
}

func (s ExpenseService) buildExpensePostingLines(ctx context.Context, organizationID uint64, lines []*ExpenseLine) ([]accounting.PostingLine, float64, error) {
	postLines := make([]accounting.PostingLine, 0, len(lines)+1)
	taxByAccount := map[uint64]float64{}
	total := 0.0
	for _, line := range lines {
		if line.CategoryID == nil {
			return nil, 0, ErrExpenseNoCategory
		}
		category, err := s.categories.Find(ctx, *line.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		if category == nil {
			return nil, 0, ErrExpenseNoCategory
		}
		if category.ExpenseAccountID == nil {
			return nil, 0, ErrExpenseNoAccount
		}
		name := ""
		if line.Description != nil {
			name = *line.Description
		}
		postLines = append(postLines, accounting.PostingLine{
			AccountID:   *category.ExpenseAccountID,
			DimensionID: line.DimensionID,
			Name:        name,
			Debit:       amount.FromFloat64(line.Amount),
		})
		total += line.Amount
		lineTaxes, err := s.lineTaxes(ctx, organizationID, line)
		if err != nil {
			return nil, 0, err
		}
		for accountID, taxAmount := range lineTaxes {
			taxByAccount[accountID] += taxAmount
			total += taxAmount
		}
	}
	for accountID, taxAmount := range taxByAccount {
		postLines = append(postLines, accounting.PostingLine{AccountID: accountID, Name: "Expense tax", Debit: amount.FromFloat64(taxAmount)})
	}
	return postLines, total, nil
}

func (s ExpenseService) lineTaxes(ctx context.Context, organizationID uint64, line *ExpenseLine) (map[uint64]float64, error) {
	taxByAccount := map[uint64]float64{}
	for _, taxID := range line.TaxIDs {
		tax, err := s.taxes.Find(ctx, uint64(taxID))
		if err != nil {
			return nil, err
		}
		if tax == nil || tax.TaxAccountID == nil {
			continue
		}
		taxAmount, err := reference.TaxLineAmount(amount.FromFloat64(line.Amount), amount.FromFloat64(line.Quantity), tax)
		if err != nil {
			return nil, err
		}
		taxByAccount[*tax.TaxAccountID] += taxAmount.Float64()
	}
	return taxByAccount, nil
}

func (s ExpenseService) payableAccount(ctx context.Context, report *ExpenseReport) (uint64, error) {
	var accountID uint64
	var err error
	if report.PaymentMode == ExpensePaymentOwnAccount {
		accountID, err = s.configs.EmployeePayableAccountID(ctx, *report.OrganizationID)
	} else {
		accountID, err = s.configs.CardClearingAccountID(ctx, *report.OrganizationID)
	}
	if err != nil {
		return 0, err
	}
	if accountID == 0 {
		return 0, ErrExpenseConfig
	}
	return accountID, nil
}

func (s ExpenseService) findReport(ctx context.Context, organizationID, reportID uint64) (*ExpenseReport, error) {
	report, err := s.reports.Search(ctx, "id", reportID)
	if err != nil {
		return nil, err
	}
	if report == nil || report.OrganizationID == nil || *report.OrganizationID != organizationID {
		return nil, ErrExpenseReportNotFound
	}
	return report, nil
}

func (s ExpenseService) ptrTime(value time.Time) *time.Time {
	return &value
}
