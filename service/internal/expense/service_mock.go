package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"gorm.io/gorm"
)

type PosterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m PosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m PosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m PosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m PosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

type ExpenseConfigSourceMock struct {
	JournalIDFunc                  func(ctx context.Context, organizationID uint64) (uint64, error)
	EmployeePayableAccountIDFunc   func(ctx context.Context, organizationID uint64) (uint64, error)
	CardClearingAccountIDFunc      func(ctx context.Context, organizationID uint64) (uint64, error)
	ReimbursementBankAccountIDFunc func(ctx context.Context, organizationID uint64) (uint64, error)
}

func (m ExpenseConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.JournalIDFunc != nil {
		return m.JournalIDFunc(ctx, organizationID)
	}
	return 10, nil
}

func (m ExpenseConfigSourceMock) EmployeePayableAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.EmployeePayableAccountIDFunc != nil {
		return m.EmployeePayableAccountIDFunc(ctx, organizationID)
	}
	return 300, nil
}

func (m ExpenseConfigSourceMock) CardClearingAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.CardClearingAccountIDFunc != nil {
		return m.CardClearingAccountIDFunc(ctx, organizationID)
	}
	return 301, nil
}

func (m ExpenseConfigSourceMock) ReimbursementBankAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.ReimbursementBankAccountIDFunc != nil {
		return m.ReimbursementBankAccountIDFunc(ctx, organizationID)
	}
	return 400, nil
}

type InvoiceEngineMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m InvoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{ContactID: request.ContactID}, nil
}

type IncomeAccountResolverMock struct {
	ResolveIncomeAccountFunc func(ctx context.Context, variantID uint64) (uint64, error)
}

func (m IncomeAccountResolverMock) ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error) {
	if m.ResolveIncomeAccountFunc != nil {
		return m.ResolveIncomeAccountFunc(ctx, variantID)
	}
	return 500, nil
}

type ProjectLookupMock struct {
	SearchFunc func(ctx context.Context, field string, value any) (*project.Project, error)
}

func (m ProjectLookupMock) Search(ctx context.Context, field string, value any) (*project.Project, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return &project.Project{ContactID: 7}, nil
}
