package expense

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestExpenseService_Create(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		req     CreateExpenseReportRequest
		wantErr error
		setup   func(captured *capturedCreate) ExpenseService
		check   func(t *testing.T, report *ExpenseReport, cap *capturedCreate)
	}{
		{
			name: "computes amounts and persists",
			setup: func(cap *capturedCreate) ExpenseService {
				return NewExpenseService(
					ExpenseReportDAOMock{
						CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
							report.ID = 7
							return report, nil
						},
						UpdateTxFunc: func(_ context.Context, _ *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
							cap.report = report
							return report, nil
						},
					},
					ExpenseLineDAOMock{
						CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *ExpenseLine) (*ExpenseLine, error) {
							cap.lines = append(cap.lines, line)
							return line, nil
						},
					},
					dao.CRUDMock[reference.ExpenseCategory]{},
					dao.CRUDMock[reference.Tax]{},
					expensePosterMock{},
					expenseConfigSourceMock{},
					expenseTransactionerMock{},
				)
			},
			req: CreateExpenseReportRequest{
				OrganizationID: 10,
				Name:           "Travel March",
				EmployeeID:     3,
				PaymentMode:    ExpensePaymentOwnAccount,
				Lines: []CreateExpenseLineRequest{
					{CategoryID: helper.Ptr(uint64(9)), Description: "Taxi", ExpenseDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Quantity: 1, UnitPrice: 50},
					{CategoryID: helper.Ptr(uint64(9)), Description: "Hotel", ExpenseDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Quantity: 2, UnitPrice: 100},
				},
			},
			wantErr: nil,
			check: func(t *testing.T, report *ExpenseReport, cap *capturedCreate) {
				if report.ID != 7 || report.State != ExpenseStateDraft {
					t.Errorf("report = %+v, want draft with id 7", report)
				}
				if len(cap.lines) != 2 {
					t.Fatalf("lines = %d, want 2", len(cap.lines))
				}
				if cap.lines[0].Amount != 50 || cap.lines[1].Amount != 200 {
					t.Errorf("line amounts = %v/%v, want 50/200", cap.lines[0].Amount, cap.lines[1].Amount)
				}
				if cap.report.TotalAmount != 250 {
					t.Errorf("total = %v, want 250", cap.report.TotalAmount)
				}
			},
		},
		{
			name: "requires lines",
			setup: func(_ *capturedCreate) ExpenseService {
				return NewExpenseService(ExpenseReportDAOMock{}, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			req:     CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: ExpensePaymentOwnAccount, Lines: []CreateExpenseLineRequest{}},
			wantErr: ErrExpenseNoLines,
		},
		{
			name: "rejects invalid payment mode",
			setup: func(_ *capturedCreate) ExpenseService {
				return NewExpenseService(ExpenseReportDAOMock{}, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			req:     CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: "card", Lines: []CreateExpenseLineRequest{{CategoryID: helper.Ptr(uint64(9))}}},
			wantErr: ErrExpenseInvalidLine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cap capturedCreate
			svc := tt.setup(&cap)
			report, err := svc.Create(ctx, tt.req)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if tt.check != nil {
				tt.check(t, report, &cap)
			}
		})
	}
}

type capturedCreate struct {
	lines  []*ExpenseLine
	report *ExpenseReport
}

func TestExpenseService_Submit(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(t *testing.T) ExpenseService
		want    string
		wantErr error
	}{
		{
			name: "transitions from draft to submitted",
			setup: func(t *testing.T) ExpenseService {
				state := ExpenseStateDraft
				reports := ExpenseReportDAOMock{
					CRUDMock: dao.CRUDMock[ExpenseReport]{
						UpdateFunc: func(_ context.Context, report *ExpenseReport) (*ExpenseReport, error) {
							state = report.State
							return report, nil
						},
					},
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: state}, nil
					},
				}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			want:    ExpenseStateSubmitted,
			wantErr: nil,
		},
		{
			name: "rejects submit when not in draft",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					CRUDMock: dao.CRUDMock[ExpenseReport]{
						UpdateFunc: func(_ context.Context, report *ExpenseReport) (*ExpenseReport, error) {
							return report, nil
						},
					},
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateSubmitted}, nil
					},
				}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			wantErr: ErrExpenseReportState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			updated, err := svc.Submit(ctx, 10, 7)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if updated.State != tt.want {
				t.Errorf("state = %s, want %s", updated.State, tt.want)
			}
			if updated.SubmittedAt == nil {
				t.Errorf("submittedAt = nil, want non-nil")
			}
		})
	}
}

func TestExpenseService_Approve(t *testing.T) {
	ctx := context.Background()
	reports := ExpenseReportDAOMock{
		CRUDMock: dao.CRUDMock[ExpenseReport]{
			UpdateFunc: func(_ context.Context, report *ExpenseReport) (*ExpenseReport, error) {
				return report, nil
			},
		},
		SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
			return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateSubmitted, PaymentMode: ExpensePaymentOwnAccount}, nil
		},
	}
	svc := NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})

	t.Run("transitions from submitted to approved", func(t *testing.T) {
		approved, err := svc.Approve(ctx, 10, 7, 5)
		helper.AssertError(t, err, false, nil)
		if approved.State != ExpenseStateApproved {
			t.Errorf("state = %s, want %s", approved.State, ExpenseStateApproved)
		}
		if approved.ApprovedBy == nil || *approved.ApprovedBy != 5 {
			t.Errorf("approvedBy = %v, want 5", approved.ApprovedBy)
		}
	})
}

func TestExpenseService_Refuse(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(t *testing.T) ExpenseService
		wantErr error
	}{
		{
			name: "rejects refuse when not in submitted state",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved}, nil
					},
				}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			wantErr: ErrExpenseReportState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			err := mustRefuse(ctx, svc)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
		})
	}
}

func TestExpenseService_Post(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(t *testing.T) ExpenseService
		wantErr error
		check   func(t *testing.T, report *ExpenseReport, posted accounting.PostRequest)
	}{
		{
			name: "own account debits expense and tax credits payable",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved, PaymentMode: ExpensePaymentOwnAccount, TotalAmount: 110}, nil
					},
				}
				lines := ExpenseLineDAOMock{
					ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
						return []*ExpenseLine{{
							Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(9)), Description: helper.Ptr("Taxi"),
							Quantity: 1, UnitPrice: 100, Amount: 100, TaxIDs: helper.Int64Array{2},
						}}, nil
					},
				}
				categories := dao.CRUDMock[reference.ExpenseCategory]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
						return &reference.ExpenseCategory{ExpenseAccountID: helper.Ptr(uint64(600))}, nil
					},
				}
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						rate := 10.0
						return &reference.Tax{Amount: &rate, Type: reference.TaxTypePercent, TaxAccountID: helper.Ptr(uint64(2200))}, nil
					},
				}
				configs := expenseConfigSourceMock{journal: 3, payable: 500}
				return NewExpenseService(reports, lines, categories, taxes, expensePosterMock{}, configs, expenseTransactionerMock{})
			},
			wantErr: nil,
			check: func(t *testing.T, report *ExpenseReport, posted accounting.PostRequest) {
				if report.State != ExpenseStatePosted || report.EntryID == nil || *report.EntryID != 99 {
					t.Errorf("report = %+v, want posted with movement 99", report)
				}
				if posted.JournalID != 3 || len(posted.Lines) != 3 {
					t.Fatalf("posted = %+v, want 3 lines on journal 3", posted)
				}
				if posted.Lines[0].AccountID != 600 || !posted.Lines[0].Debit.Equal(amount.FromFloat64(100)) {
					t.Errorf("expense line = %+v, want Dr 600 of 100", posted.Lines[0])
				}
				if posted.Lines[1].AccountID != 2200 || !posted.Lines[1].Debit.Equal(amount.FromFloat64(10)) {
					t.Errorf("tax line = %+v, want Dr 2200 of 10", posted.Lines[1])
				}
				if posted.Lines[2].AccountID != 500 || !posted.Lines[2].Credit.Equal(amount.FromFloat64(110)) {
					t.Errorf("payable line = %+v, want Cr 500 of 110", posted.Lines[2])
				}
			},
		},
		{
			name: "org account credits card clearing",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved, PaymentMode: ExpensePaymentOrgAccount}, nil
					},
				}
				lines := ExpenseLineDAOMock{
					ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
						return []*ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(9)), Quantity: 1, UnitPrice: 80, Amount: 80}}, nil
					},
				}
				categories := dao.CRUDMock[reference.ExpenseCategory]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
						return &reference.ExpenseCategory{ExpenseAccountID: helper.Ptr(uint64(600))}, nil
					},
				}
				configs := expenseConfigSourceMock{journal: 3, clearing: 700}
				return NewExpenseService(reports, lines, categories, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, configs, expenseTransactionerMock{})
			},
			wantErr: nil,
			check: func(t *testing.T, report *ExpenseReport, posted accounting.PostRequest) {
				if posted.Lines[len(posted.Lines)-1].AccountID != 700 {
					t.Errorf("clearing line = %+v, want Cr 700", posted.Lines[len(posted.Lines)-1])
				}
			},
		},
		{
			name: "requires approved state",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateDraft, PaymentMode: ExpensePaymentOwnAccount}, nil
					},
				}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			wantErr: ErrExpenseReportState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			var posted accounting.PostRequest
			svc.poster = expensePosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			}
			report, err := svc.Post(ctx, 10, 7)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if tt.check != nil {
				tt.check(t, report, posted)
			}
		})
	}
}

func TestExpenseService_Reimburse(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(t *testing.T) ExpenseService
		wantErr error
		check   func(t *testing.T, report *ExpenseReport, posted accounting.PostRequest)
	}{
		{
			name: "clears payable to bank",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, PaymentMode: ExpensePaymentOwnAccount, TotalAmount: 110}, nil
					},
				}
				configs := expenseConfigSourceMock{journal: 3, payable: 500, bank: 400}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, configs, expenseTransactionerMock{})
			},
			wantErr: nil,
			check: func(t *testing.T, report *ExpenseReport, posted accounting.PostRequest) {
				if report.State != ExpenseStateReimbursed {
					t.Errorf("state = %s, want %s", report.State, ExpenseStateReimbursed)
				}
				if report.ReimbursementEntryID == nil || *report.ReimbursementEntryID != 99 {
					t.Errorf("reimbursementMovementID = %v, want 99", report.ReimbursementEntryID)
				}
				if len(posted.Lines) != 2 || posted.Lines[0].AccountID != 500 || posted.Lines[1].AccountID != 400 {
					t.Errorf("posted = %+v, want Dr 500 / Cr 400", posted.Lines)
				}
			},
		},
		{
			name: "only own account",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, PaymentMode: ExpensePaymentOrgAccount}, nil
					},
				}
				return NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})
			},
			wantErr: ErrExpenseReportState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			var posted accounting.PostRequest
			svc.poster = expensePosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			}
			report, err := svc.Reimburse(ctx, 10, 7)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if tt.check != nil {
				tt.check(t, report, posted)
			}
		})
	}
}

func TestExpenseService_BillToInvoice(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(t *testing.T) ExpenseService
		wantErr error
		check   func(t *testing.T, invoiceRequest accounting.CreateInvoiceRequest)
	}{
		{
			name: "creates customer invoice for billable lines",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, Name: "Travel", PaymentMode: ExpensePaymentOwnAccount}, nil
					},
				}
				lines := ExpenseLineDAOMock{
					ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
						return []*ExpenseLine{
							{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(11)), Quantity: 2, UnitPrice: 100, ProjectID: helper.Ptr(uint64(31)), Reimbursable: false, Description: helper.Ptr("Site visit")},
							{Base: model.Base{ID: 2}, CategoryID: helper.Ptr(uint64(9)), Quantity: 1, UnitPrice: 30, Reimbursable: true},
						}, nil
					},
				}
				invoices := invoiceEngineMock{
					createFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: model.Base{ID: 55}}, nil
					},
				}
				products := incomeAccountResolverMock{account: 4100}
				projects := projectLookupMock{
					searchFunc: func(_ context.Context, _ string, _ any) (*project.Project, error) {
						return &project.Project{ContactID: 42}, nil
					},
				}
				configs := expenseConfigSourceMock{journal: 3}
				return NewExpenseService(reports, lines, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, configs, expenseTransactionerMock{}).SetBilling(invoices, products, projects)
			},
			wantErr: nil,
			check: func(t *testing.T, invoiceRequest accounting.CreateInvoiceRequest) {
				if invoiceRequest.ContactID != 42 {
					t.Errorf("contact = %d, want 42", invoiceRequest.ContactID)
				}
				if len(invoiceRequest.Lines) != 1 || invoiceRequest.Lines[0].AccountID != 4100 {
					t.Errorf("invoice lines = %+v, want one billable line on 4100", invoiceRequest.Lines)
				}
			},
		},
		{
			name: "requires billable lines",
			setup: func(t *testing.T) ExpenseService {
				reports := ExpenseReportDAOMock{
					SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
						return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, PaymentMode: ExpensePaymentOwnAccount}, nil
					},
				}
				lines := ExpenseLineDAOMock{
					ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
						return []*ExpenseLine{{Base: model.Base{ID: 2}, CategoryID: helper.Ptr(uint64(9)), Reimbursable: true}}, nil
					},
				}
				return NewExpenseService(reports, lines, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{}).SetBilling(invoiceEngineMock{}, incomeAccountResolverMock{}, projectLookupMock{})
			},
			wantErr: ErrExpenseNoBillable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			var invoiceRequest accounting.CreateInvoiceRequest
			origMock := svc.invoices.(invoiceEngineMock)
			svc.invoices = invoiceEngineMock{
				createFunc: func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
					invoiceRequest = request
					if origMock.createFunc != nil {
						return origMock.createFunc(ctx, request)
					}
					return &accounting.Invoice{Base: model.Base{ID: 55}}, nil
				},
			}
			_, err := svc.BillToInvoice(ctx, 10, 7)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			if tt.check != nil {
				tt.check(t, invoiceRequest)
			}
		})
	}
}

func TestExpenseService_Get(t *testing.T) {
	ctx := context.Background()
	reports := ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
			return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99)), State: ExpenseStateDraft}, nil
		},
	}
	svc := NewExpenseService(reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, expensePosterMock{}, expenseConfigSourceMock{}, expenseTransactionerMock{})

	t.Run("scopes to organization", func(t *testing.T) {
		err := mustGetReport(ctx, svc)
		helper.AssertError(t, err, true, ErrExpenseReportNotFound)
	})
}

func mustSubmit(ctx context.Context, svc ExpenseService) (err error) {
	_, err = svc.Submit(ctx, 10, 7)
	return err
}

func mustRefuse(ctx context.Context, svc ExpenseService) (err error) {
	_, err = svc.Refuse(ctx, 10, 7)
	return err
}

func mustGetReport(ctx context.Context, svc ExpenseService) (err error) {
	_, err = svc.Get(ctx, 10, 7)
	return err
}

type expensePosterMock struct {
	PostTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
}

func (m expensePosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (m expensePosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
}

func (m expensePosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (m expensePosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return m.Reverse(ctx, request)
}

type expenseTransactionerMock struct{}

func (m expenseTransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return fn(nil)
}

var _ db.Transactioner = expenseTransactionerMock{}

type expenseConfigSourceMock struct {
	journal  uint64
	payable  uint64
	clearing uint64
	bank     uint64
}

func (m expenseConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	return m.journal, nil
}

func (m expenseConfigSourceMock) EmployeePayableAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	return m.payable, nil
}

func (m expenseConfigSourceMock) CardClearingAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	return m.clearing, nil
}

func (m expenseConfigSourceMock) ReimbursementBankAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	return m.bank, nil
}

var _ ExpenseConfigSource = expenseConfigSourceMock{}

type incomeAccountResolverMock struct {
	account uint64
}

func (m incomeAccountResolverMock) ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error) {
	return m.account, nil
}

type invoiceEngineMock struct {
	createFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m invoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
}

type projectLookupMock struct {
	searchFunc func(ctx context.Context, field string, value any) (*project.Project, error)
}

func (m projectLookupMock) Search(ctx context.Context, field string, value any) (*project.Project, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, field, value)
	}
	return &project.Project{}, nil
}
