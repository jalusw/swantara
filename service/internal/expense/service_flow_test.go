package expense

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestList(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		reports ExpenseReportDAOMock
		wantLen int
		wantIDs []uint64
		wantErr bool
	}{
		{
			name: "returns organization reports",
			reports: ExpenseReportDAOMock{
				CRUDMock: dao.CRUDMock[ExpenseReport]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[ExpenseReport], error) {
						return &query.Page[ExpenseReport]{Items: []*ExpenseReport{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))},
							{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(10))},
						}}, nil
					},
				},
			},
			wantLen: 2,
			wantIDs: []uint64{1, 2},
		},
		{
			name: "propagates error",
			reports: ExpenseReportDAOMock{
				CRUDMock: dao.CRUDMock[ExpenseReport]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[ExpenseReport], error) {
						return nil, errors.New("db down")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			items, err := svc.List(ctx, 10)
			if tc.wantErr {
				helper.AssertError(t, err, true, nil)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tc.wantLen {
				t.Errorf("items = %+v, want %d items", items, tc.wantLen)
				return
			}
			for i, id := range tc.wantIDs {
				if items[i].ID != id {
					t.Errorf("items[%d].ID = %d, want %d", i, items[i].ID, id)
				}
			}
		})
	}
}

func TestListLines(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		lines         ExpenseLineDAOMock
		wantLen       int
		wantIDs       []uint64
		wantErr       bool
		wantErrTarget error
	}{
		{
			name: "returns report lines",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}}, nil
				},
			},
			wantLen: 2,
			wantIDs: []uint64{1, 2},
		},
		{
			name: "returns empty with default mock",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
				},
			},
			lines:   ExpenseLineDAOMock{},
			wantLen: 0,
		},
		{
			name: "scopes to organization",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
				},
			},
			lines:         ExpenseLineDAOMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseReportNotFound,
		},
		{
			name: "propagates lines error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, tc.lines, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			items, err := svc.ListLines(ctx, 10, 7)
			if tc.wantErr {
				helper.AssertError(t, err, true, tc.wantErrTarget)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tc.wantLen {
				t.Errorf("items = %+v, want %d items", items, tc.wantLen)
				return
			}
			for i, id := range tc.wantIDs {
				if items[i].ID != id {
					t.Errorf("items[%d].ID = %d, want %d", i, items[i].ID, id)
				}
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name      string
		reports   ExpenseReportDAOMock
		lines     ExpenseLineDAOMock
		tx        TransactionerMock
		request   CreateExpenseReportRequest
		wantErr   bool
		wantTotal float64
	}{
		{
			name:    "rejects invalid payment mode",
			reports: ExpenseReportDAOMock{},
			lines:   ExpenseLineDAOMock{},
			request: CreateExpenseReportRequest{
				OrganizationID: 10, PaymentMode: "card", Lines: []CreateExpenseLineRequest{{Description: "Taxi"}},
			},
			wantErr: true,
		},
		{
			name: "propagates report create error",
			reports: ExpenseReportDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ExpenseReport) (*ExpenseReport, error) {
					return nil, errors.New("insert failed")
				},
			},
			lines:   ExpenseLineDAOMock{},
			request: CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name: "propagates line create error",
			reports: ExpenseReportDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
					report.ID = 7
					return report, nil
				},
			},
			lines: ExpenseLineDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ExpenseLine) (*ExpenseLine, error) {
					return nil, errors.New("insert failed")
				},
			},
			request: CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name: "propagates report update error",
			reports: ExpenseReportDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
					report.ID = 7
					return report, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ExpenseReport) (*ExpenseReport, error) {
					return nil, errors.New("update failed")
				},
			},
			lines:   ExpenseLineDAOMock{},
			request: CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name:    "propagates transaction error",
			reports: ExpenseReportDAOMock{},
			lines:   ExpenseLineDAOMock{},
			tx: TransactionerMock{
				RunFunc: func(_ context.Context, _ func(tx *gorm.DB) error) error {
					return errors.New("tx down")
				},
			},
			request: CreateExpenseReportRequest{OrganizationID: 10, PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name:    "computes total with default mocks",
			reports: ExpenseReportDAOMock{},
			lines:   ExpenseLineDAOMock{},
			request: CreateExpenseReportRequest{
				OrganizationID: 10,
				Name:           "Travel",
				PaymentMode:    ExpensePaymentOwnAccount,
				Lines: []CreateExpenseLineRequest{
					{Description: "Taxi", Quantity: 1, UnitPrice: 50},
					{Description: "Hotel", Quantity: 2, UnitPrice: 100},
				},
			},
			wantTotal: 250,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, tc.lines, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, tc.tx)
			report, err := svc.Create(ctx, tc.request)
			if tc.wantErr {
				helper.AssertError(t, err, true, nil)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantTotal > 0 && report.TotalAmount != tc.wantTotal {
				t.Errorf("total = %v, want %v", report.TotalAmount, tc.wantTotal)
			}
		})
	}
}

func TestSubmit(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		reports ExpenseReportDAOMock
		wantErr bool
	}{
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
		{
			name: "propagates update error",
			reports: ExpenseReportDAOMock{
				CRUDMock: dao.CRUDMock[ExpenseReport]{
					UpdateFunc: func(_ context.Context, _ *ExpenseReport) (*ExpenseReport, error) {
						return nil, errors.New("update failed")
					},
				},
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateDraft}, nil
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			helper.AssertError(t, mustSubmit(ctx, svc), true, nil)
		})
	}
}

func TestApprove(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		wantErr       bool
		wantErrTarget error
	}{
		{
			name: "requires submitted state",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved}, nil
				},
			},
			wantErr:       true,
			wantErrTarget: ErrExpenseReportState,
		},
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
		{
			name: "propagates update error",
			reports: ExpenseReportDAOMock{
				CRUDMock: dao.CRUDMock[ExpenseReport]{
					UpdateFunc: func(_ context.Context, _ *ExpenseReport) (*ExpenseReport, error) {
						return nil, errors.New("update failed")
					},
				},
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateSubmitted}, nil
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			helper.AssertError(t, mustApprove(ctx, svc), true, tc.wantErrTarget)
		})
	}
}

func TestRefuse(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		wantErr       bool
		wantErrTarget error
	}{
		{
			name: "refuses submitted report",
			reports: ExpenseReportDAOMock{
				CRUDMock: dao.CRUDMock[ExpenseReport]{
					UpdateFunc: func(_ context.Context, report *ExpenseReport) (*ExpenseReport, error) {
						return report, nil
					},
				},
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateSubmitted}, nil
				},
			},
		},
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			refused, err := svc.Refuse(ctx, 10, 7)
			if tc.wantErr {
				helper.AssertError(t, err, true, tc.wantErrTarget)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if refused.State != ExpenseStateRefused {
				t.Errorf("refused.State = %v, want %v", refused.State, ExpenseStateRefused)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		wantErr       bool
		wantErrTarget error
	}{
		{
			name:          "not found with default mock",
			reports:       ExpenseReportDAOMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseReportNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			helper.AssertError(t, mustGetReport(ctx, svc), true, tc.wantErrTarget)
		})
	}
}

func TestFindReport(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		reports ExpenseReportDAOMock
		wantErr bool
	}{
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			helper.AssertError(t, mustFindReport(ctx, svc), true, nil)
		})
	}
}

func TestPost(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name       string
		reports    ExpenseReportDAOMock
		lines      ExpenseLineDAOMock
		categories dao.CRUDMock[reference.ExpenseCategory]
		taxes      dao.CRUDMock[reference.Tax]
		poster     PosterMock
		configs    ExpenseConfigSourceMock
		tx         TransactionerMock
		wantErr    bool
	}{
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			lines:   ExpenseLineDAOMock{},
			wantErr: true,
		},
		{
			name: "propagates lines error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
		{
			name: "requires lines",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "propagates journal error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := postReadyConfigs()
				c.JournalIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "requires journal",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := postReadyConfigs()
				c.JournalIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "propagates payable error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := postReadyConfigs()
				c.EmployeePayableAccountIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "requires payable account",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := postReadyConfigs()
				c.EmployeePayableAccountIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "propagates category error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return nil, errors.New("db down")
				},
			},
			poster:  PosterMock{},
			configs: postReadyConfigs(),
			wantErr: true,
		},
		{
			name: "propagates posting error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: postReadyCategories(),
			poster: PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, errors.New("post failed")
				},
			},
			configs: postReadyConfigs(),
			wantErr: true,
		},
		{
			name: "propagates report update error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ExpenseReport) (*ExpenseReport, error) {
					return nil, errors.New("update failed")
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: postReadyCategories(),
			poster: PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			},
			configs: postReadyConfigs(),
			wantErr: true,
		},
		{
			name: "propagates transaction error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: postReadyCategories(),
			configs:    postReadyConfigs(),
			tx: TransactionerMock{
				RunFunc: func(_ context.Context, _ func(tx *gorm.DB) error) error {
					return errors.New("tx down")
				},
			},
			wantErr: true,
		},
		{
			name: "uses default config for org payment",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOrgAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: postReadyCategories(),
		},
		{
			name: "posts with default post tx",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return approvedReport(ExpensePaymentOwnAccount), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return postReadyLines(), nil
				},
			},
			categories: postReadyCategories(),
			poster: PosterMock{
				PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			},
			configs: postReadyConfigs(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, tc.lines, tc.categories, tc.taxes, tc.poster, tc.configs, tc.tx)
			report, err := svc.Post(ctx, 10, 7)
			if tc.wantErr {
				helper.AssertError(t, err, true, nil)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if report.State != ExpenseStatePosted {
				t.Errorf("report.State = %v, want %v", report.State, ExpenseStatePosted)
			}
			if report.EntryID == nil {
				t.Errorf("report.EntryID = nil, want non-nil")
			}
		})
	}
}

func TestReimburse(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		poster        PosterMock
		configs       ExpenseConfigSourceMock
		tx            TransactionerMock
		wantErr       bool
		wantErrTarget error
	}{
		{
			name: "requires posted state",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved, PaymentMode: ExpensePaymentOwnAccount}, nil
				},
			},
			wantErr:       true,
			wantErrTarget: ErrExpenseReportState,
		},
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
		{
			name: "propagates journal error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := reimburseReadyConfigs()
				c.JournalIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "requires journal",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := reimburseReadyConfigs()
				c.JournalIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				}
				return c
			}(),
			wantErr:       true,
			wantErrTarget: ErrExpenseConfig,
		},
		{
			name: "propagates bank error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := reimburseReadyConfigs()
				c.ReimbursementBankAccountIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "requires bank account",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := reimburseReadyConfigs()
				c.ReimbursementBankAccountIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				}
				return c
			}(),
			wantErr:       true,
			wantErrTarget: ErrExpenseConfig,
		},
		{
			name: "propagates payable error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: func() ExpenseConfigSourceMock {
				c := reimburseReadyConfigs()
				c.EmployeePayableAccountIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "propagates posting error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			poster: PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, errors.New("post failed")
				},
			},
			configs: reimburseReadyConfigs(),
			wantErr: true,
		},
		{
			name: "propagates report update error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ExpenseReport) (*ExpenseReport, error) {
					return nil, errors.New("update failed")
				},
			},
			poster: PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			},
			configs: reimburseReadyConfigs(),
			wantErr: true,
		},
		{
			name: "propagates transaction error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
			configs: reimburseReadyConfigs(),
			tx: TransactionerMock{
				RunFunc: func(_ context.Context, _ func(tx *gorm.DB) error) error {
					return errors.New("tx down")
				},
			},
			wantErr: true,
		},
		{
			name: "uses default config",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return postedReport(), nil
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, tc.poster, tc.configs, tc.tx)
			report, err := svc.Reimburse(ctx, 10, 7)
			if tc.wantErr {
				helper.AssertError(t, err, true, tc.wantErrTarget)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if report.State != ExpenseStateReimbursed {
				t.Errorf("report.State = %v, want %v", report.State, ExpenseStateReimbursed)
			}
			if report.ReimbursementEntryID == nil {
				t.Errorf("report.ReimbursementEntryID = nil, want non-nil")
			}
		})
	}
}

func TestBillToInvoice(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		reports       ExpenseReportDAOMock
		lines         ExpenseLineDAOMock
		poster        PosterMock
		configs       ExpenseConfigSourceMock
		invoices      InvoiceEngineMock
		products      IncomeAccountResolverMock
		projects      ProjectLookupMock
		wantErr       bool
		wantErrTarget error
	}{
		{
			name:    "requires billing dependencies",
			reports: ExpenseReportDAOMock{},
			lines:   ExpenseLineDAOMock{},
			configs: ExpenseConfigSourceMock{},
			wantErr: true,
		},
		{
			name: "propagates search error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return nil, errors.New("db down")
				},
			},
			lines:   ExpenseLineDAOMock{},
			configs: ExpenseConfigSourceMock{},
			wantErr: true,
		},
		{
			name: "requires posted state",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved, Name: "Travel", PaymentMode: ExpensePaymentOwnAccount}, nil
				},
			},
			lines:         ExpenseLineDAOMock{},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseReportState,
		},
		{
			name: "propagates lines error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return nil, errors.New("db down")
				},
			},
			configs: ExpenseConfigSourceMock{},
			wantErr: true,
		},
		{
			name: "requires billable project",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(11)), Reimbursable: false}}, nil
				},
			},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseBillableProject,
		},
		{
			name: "requires single project",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{
						billableLine(31, 11),
						{Base: model.Base{ID: 2}, ItemID: helper.Ptr(uint64(12)), ProjectID: helper.Ptr(uint64(32)), Reimbursable: false},
					}, nil
				},
			},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseBillableProject,
		},
		{
			name: "propagates project error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			projects: ProjectLookupMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*project.Project, error) {
					return nil, errors.New("db down")
				},
			},
			configs: ExpenseConfigSourceMock{},
			wantErr: true,
		},
		{
			name: "requires found project",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			projects: ProjectLookupMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*project.Project, error) {
					return nil, nil
				},
			},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseProjectNotFound,
		},
		{
			name: "propagates journal error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			configs: ExpenseConfigSourceMock{
				JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				},
			},
			wantErr: true,
		},
		{
			name: "requires journal",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			configs: ExpenseConfigSourceMock{
				JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				},
			},
			wantErr:       true,
			wantErrTarget: ErrExpenseConfig,
		},
		{
			name: "requires income account",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{{Base: model.Base{ID: 1}, ProjectID: helper.Ptr(uint64(31)), Reimbursable: false}}, nil
				},
			},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseNoIncomeAccount,
		},
		{
			name: "propagates income account error",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			products: IncomeAccountResolverMock{
				ResolveIncomeAccountFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("db down")
				},
			},
			configs: ExpenseConfigSourceMock{},
			wantErr: true,
		},
		{
			name: "requires resolved income account",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			products: IncomeAccountResolverMock{
				ResolveIncomeAccountFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				},
			},
			configs:       ExpenseConfigSourceMock{},
			wantErr:       true,
			wantErrTarget: ErrExpenseNoIncomeAccount,
		},
		{
			name: "uses default dependencies",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			configs: ExpenseConfigSourceMock{},
		},
		{
			name: "delegates to invoice engine",
			reports: ExpenseReportDAOMock{
				SearchFunc: func(_ context.Context, _ string, _ any) (*ExpenseReport, error) {
					return billableReport(), nil
				},
			},
			lines: ExpenseLineDAOMock{
				ListByReportFunc: func(_ context.Context, _ uint64) ([]*ExpenseLine, error) {
					return []*ExpenseLine{billableLine(31, 11)}, nil
				},
			},
			invoices: InvoiceEngineMock{
				CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 55}}, nil
				},
			},
			configs: ExpenseConfigSourceMock{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(tc.reports, tc.lines, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, tc.poster, tc.configs, TransactionerMock{}).SetBilling(tc.invoices, tc.products, tc.projects)
			invoice, err := svc.BillToInvoice(ctx, 10, 7)
			if tc.wantErr {
				helper.AssertError(t, err, true, tc.wantErrTarget)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.invoices.CreateFunc != nil {
				if invoice.ID != 55 {
					t.Errorf("invoice.ID = %v, want 55", invoice.ID)
				}
			} else {
				if invoice.ContactID != 7 {
					t.Errorf("invoice.ContactID = %v, want 7", invoice.ContactID)
				}
			}
		})
	}
}

func TestBuildExpensePostingLines(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		categories    dao.CRUDMock[reference.ExpenseCategory]
		taxes         dao.CRUDMock[reference.Tax]
		lines         []*ExpenseLine
		wantErr       bool
		wantErrTarget error
		wantLines     int
		wantTotal     float64
	}{
		{
			name:    "requires category",
			lines:   []*ExpenseLine{{Amount: 100}},
			wantErr: true,
		},
		{
			name: "propagates category error",
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return nil, errors.New("db down")
				},
			},
			lines:   []*ExpenseLine{{CategoryID: helper.Ptr(uint64(9)), Amount: 100}},
			wantErr: true,
		},
		{
			name: "requires found category",
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return nil, nil
				},
			},
			lines:   []*ExpenseLine{{CategoryID: helper.Ptr(uint64(9)), Amount: 100}},
			wantErr: true,
		},
		{
			name: "requires expense account",
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return &reference.ExpenseCategory{}, nil
				},
			},
			lines:         []*ExpenseLine{{CategoryID: helper.Ptr(uint64(9)), Amount: 100}},
			wantErr:       true,
			wantErrTarget: ErrExpenseNoAccount,
		},
		{
			name: "propagates tax error",
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return &reference.ExpenseCategory{ExpenseAccountID: helper.Ptr(uint64(600))}, nil
				},
			},
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					return nil, errors.New("db down")
				},
			},
			lines:   []*ExpenseLine{{CategoryID: helper.Ptr(uint64(9)), Amount: 100, TaxIDs: helper.Int64Array{2}}},
			wantErr: true,
		},
		{
			name: "accumulates taxes",
			categories: dao.CRUDMock[reference.ExpenseCategory]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
					return &reference.ExpenseCategory{ExpenseAccountID: helper.Ptr(uint64(600))}, nil
				},
			},
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					rate := 10.0
					return &reference.Tax{Amount: &rate, Type: reference.TaxTypePercent, TaxAccountID: helper.Ptr(uint64(2200))}, nil
				},
			},
			lines: []*ExpenseLine{
				{CategoryID: helper.Ptr(uint64(9)), Description: helper.Ptr("Taxi"), Quantity: 1, UnitPrice: 100, Amount: 100, TaxIDs: helper.Int64Array{2}},
			},
			wantLines: 2,
			wantTotal: 110,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(ExpenseReportDAOMock{}, ExpenseLineDAOMock{}, tc.categories, tc.taxes, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			lines, total, err := svc.buildExpensePostingLines(ctx, 1, tc.lines)
			if tc.wantErr {
				helper.AssertError(t, err, true, tc.wantErrTarget)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(lines) != tc.wantLines || total != tc.wantTotal {
				t.Errorf("lines = %+v, total = %v, want %d lines totalling %v", lines, total, tc.wantLines, tc.wantTotal)
			}
		})
	}
}

func TestLineTaxes(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		taxes   dao.CRUDMock[reference.Tax]
		line    *ExpenseLine
		wantErr bool
		wantLen int
	}{
		{
			name: "propagates tax error",
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					return nil, errors.New("db down")
				},
			},
			line:    &ExpenseLine{Amount: 100, TaxIDs: helper.Int64Array{2}},
			wantErr: true,
		},
		{
			name: "skips tax without account",
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					rate := 10.0
					return &reference.Tax{Amount: &rate, Type: reference.TaxTypePercent}, nil
				},
			},
			line:    &ExpenseLine{Amount: 100, TaxIDs: helper.Int64Array{2}},
			wantLen: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(ExpenseReportDAOMock{}, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, tc.taxes, PosterMock{}, ExpenseConfigSourceMock{}, TransactionerMock{})
			taxByAccount, err := svc.lineTaxes(ctx, 1, tc.line)
			if tc.wantErr {
				helper.AssertError(t, err, true, nil)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(taxByAccount) != tc.wantLen {
				t.Errorf("taxByAccount = %+v, want %d items", taxByAccount, tc.wantLen)
			}
		})
	}
}

func TestPayableAccount(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name        string
		configs     ExpenseConfigSourceMock
		report      *ExpenseReport
		wantErr     bool
		wantAccount uint64
	}{
		{
			name: "propagates config error",
			configs: ExpenseConfigSourceMock{
				EmployeePayableAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, errors.New("config down")
				},
			},
			report:  &ExpenseReport{OrganizationID: helper.Ptr(uint64(10)), PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name: "requires account",
			configs: ExpenseConfigSourceMock{
				EmployeePayableAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 0, nil
				},
			},
			report:  &ExpenseReport{OrganizationID: helper.Ptr(uint64(10)), PaymentMode: ExpensePaymentOwnAccount},
			wantErr: true,
		},
		{
			name: "reads card clearing account",
			configs: ExpenseConfigSourceMock{
				CardClearingAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
					return 700, nil
				},
			},
			report:      &ExpenseReport{OrganizationID: helper.Ptr(uint64(10)), PaymentMode: ExpensePaymentOrgAccount},
			wantAccount: 700,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewExpenseService(ExpenseReportDAOMock{}, ExpenseLineDAOMock{}, dao.CRUDMock[reference.ExpenseCategory]{}, dao.CRUDMock[reference.Tax]{}, PosterMock{}, tc.configs, TransactionerMock{})
			account, err := svc.payableAccount(ctx, tc.report)
			if tc.wantErr {
				helper.AssertError(t, err, true, nil)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if account != tc.wantAccount {
				t.Errorf("account = %d, want %d", account, tc.wantAccount)
			}
		})
	}
}

func TestPosterMock_ReverseDefaults(t *testing.T) {
	ctx := context.Background()
	var poster PosterMock

	if _, err := poster.Reverse(ctx, accounting.ReverseRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := poster.ReverseTx(ctx, nil, accounting.ReverseRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPosterMock_ReverseDelegates(t *testing.T) {
	ctx := context.Background()
	reverseCalled := false
	reverseTxCalled := false
	poster := PosterMock{
		ReverseFunc: func(_ context.Context, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
			reverseCalled = true
			return &accounting.JournalEntry{}, nil
		},
		ReverseTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
			reverseTxCalled = true
			return &accounting.JournalEntry{}, nil
		},
	}

	if _, err := poster.Reverse(ctx, accounting.ReverseRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := poster.ReverseTx(ctx, nil, accounting.ReverseRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reverseCalled || !reverseTxCalled {
		t.Errorf("reverse = %v, reverseTx = %v, want both delegated", reverseCalled, reverseTxCalled)
	}
}

func approvedReport(paymentMode string) *ExpenseReport {
	return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStateApproved, Name: "Travel", PaymentMode: paymentMode}
}

func postedReport() *ExpenseReport {
	return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, Name: "Travel", PaymentMode: ExpensePaymentOwnAccount, TotalAmount: 110}
}

func billableLine(projectID, itemID uint64) *ExpenseLine {
	return &ExpenseLine{Base: model.Base{ID: 1}, ItemID: helper.Ptr(itemID), ProjectID: helper.Ptr(projectID), Reimbursable: false, Description: helper.Ptr("Site visit"), Quantity: 2, UnitPrice: 100}
}

func billableReport() *ExpenseReport {
	return &ExpenseReport{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), State: ExpenseStatePosted, Name: "Travel", PaymentMode: ExpensePaymentOwnAccount}
}

func postReadyLines() []*ExpenseLine {
	return []*ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(9)), Quantity: 1, UnitPrice: 100, Amount: 100}}
}

func postReadyCategories() dao.CRUDMock[reference.ExpenseCategory] {
	return dao.CRUDMock[reference.ExpenseCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
			return &reference.ExpenseCategory{ExpenseAccountID: helper.Ptr(uint64(600))}, nil
		},
	}
}

func postReadyConfigs() ExpenseConfigSourceMock {
	return ExpenseConfigSourceMock{
		JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 3, nil
		},
		EmployeePayableAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 500, nil
		},
	}
}

func reimburseReadyConfigs() ExpenseConfigSourceMock {
	return ExpenseConfigSourceMock{
		JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 3, nil
		},
		ReimbursementBankAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 400, nil
		},
		EmployeePayableAccountIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 500, nil
		},
	}
}

func mustApprove(ctx context.Context, svc ExpenseService) (err error) {
	_, err = svc.Approve(ctx, 10, 7, 5)
	return err
}

func mustFindReport(ctx context.Context, svc ExpenseService) (err error) {
	_, err = svc.findReport(ctx, 10, 7)
	return err
}
