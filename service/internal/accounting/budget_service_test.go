package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestBudgetService_Create(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "stores budget with lines",
			fn: func(t *testing.T) {
				ctx := context.Background()
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

				var createdLines []*BudgetLine
				budgets := BudgetDAOMock{
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, budget *Budget, lines []*BudgetLine) (*Budget, error) {
						budget.ID = 1
						for _, line := range lines {
							line.BudgetID = budget.ID
						}
						createdLines = lines
						return budget, nil
					},
				}
				svc := NewBudgetService(budgets, BudgetLineDAOMock{}, BudgetQueryDAOMock{}, TransactionerMock{})

				budget, err := svc.Create(ctx, CreateBudgetRequest{
					OrganizationID: 10,
					Name:           "FY2026",
					DateStart:      start,
					DateEnd:        end,
					Lines: []BudgetLineRequest{
						{AccountID: 4100, PlannedAmount: 1000},
					},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if budget.State != BudgetStateDraft {
					t.Errorf("state = %s, want draft", budget.State)
				}
				if len(createdLines) != 1 || createdLines[0].AccountID != 4100 || createdLines[0].BudgetID != 1 {
					t.Errorf("lines = %+v, want account 4100 on budget 1", createdLines)
				}
			},
		},
		{
			name: "rejects without lines",
			fn: func(t *testing.T) {
				ctx := context.Background()
				svc := NewBudgetService(BudgetDAOMock{}, BudgetLineDAOMock{}, BudgetQueryDAOMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateBudgetRequest{OrganizationID: 10})
				if helper.AssertError(t, err, true, ErrBudgetNoLines) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestBudgetService_Variance(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "computes practical from movements",
			fn: func(t *testing.T) {
				ctx := context.Background()
				budget := &Budget{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), DateStart: &start, DateEnd: &end}
				budgets := BudgetDAOMock{
					CRUDMock: dao.CRUDMock[Budget]{
						FindFunc: func(_ context.Context, _ uint64) (*Budget, error) { return budget, nil },
					},
				}
				lines := BudgetLineDAOMock{
					ListByBudgetFunc: func(_ context.Context, _ uint64) ([]*BudgetLine, error) {
						return []*BudgetLine{{Base: model.Base{ID: 1}, AccountID: 4100, PlannedAmount: 1000}}, nil
					},
				}
				query := BudgetQueryDAOMock{
					SumPracticalByPeriodFunc: func(_ context.Context, _ uint64, _ uint64, _, _ time.Time) (float64, error) {
						return 800, nil
					},
				}
				svc := NewBudgetService(budgets, lines, query, TransactionerMock{})

				variance, err := svc.Variance(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(variance) != 1 {
					t.Fatalf("variance = %d, want 1", len(variance))
				}
				if variance[0].PlannedAmount != 1000 || variance[0].PracticalAmount != 800 || variance[0].Variance != -200 {
					t.Errorf("variance = %+v, want planned 1000 practical 800 variance -200", variance[0])
				}
			},
		},
		{
			name: "rejects missing budget",
			fn: func(t *testing.T) {
				ctx := context.Background()
				svc := NewBudgetService(BudgetDAOMock{}, BudgetLineDAOMock{}, BudgetQueryDAOMock{}, TransactionerMock{})

				_, err := svc.Variance(ctx, 99)
				if helper.AssertError(t, err, true, ErrBudgetNotFound) {
					return
				}
			},
		},
		{
			name: "propagates find error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				budgets := BudgetDAOMock{
					CRUDMock: dao.CRUDMock[Budget]{
						FindFunc: func(_ context.Context, _ uint64) (*Budget, error) { return nil, errors.New("db down") },
					},
				}
				svc := NewBudgetService(budgets, BudgetLineDAOMock{}, BudgetQueryDAOMock{}, TransactionerMock{})

				_, err := svc.Variance(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates line error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				budget := &Budget{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}
				budgets := BudgetDAOMock{
					CRUDMock: dao.CRUDMock[Budget]{
						FindFunc: func(_ context.Context, _ uint64) (*Budget, error) { return budget, nil },
					},
				}
				lines := BudgetLineDAOMock{
					ListByBudgetFunc: func(_ context.Context, _ uint64) ([]*BudgetLine, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewBudgetService(budgets, lines, BudgetQueryDAOMock{}, TransactionerMock{})

				_, err := svc.Variance(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates sum error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				budget := &Budget{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), DateStart: &start, DateEnd: &end}
				budgets := BudgetDAOMock{
					CRUDMock: dao.CRUDMock[Budget]{
						FindFunc: func(_ context.Context, _ uint64) (*Budget, error) { return budget, nil },
					},
				}
				lines := BudgetLineDAOMock{
					ListByBudgetFunc: func(_ context.Context, _ uint64) ([]*BudgetLine, error) {
						return []*BudgetLine{{Base: model.Base{ID: 1}, AccountID: 4100, PlannedAmount: 1000}}, nil
					},
				}
				query := BudgetQueryDAOMock{
					SumPracticalByPeriodFunc: func(_ context.Context, _ uint64, _ uint64, _, _ time.Time) (float64, error) {
						return 0, errors.New("db down")
					},
				}
				svc := NewBudgetService(budgets, lines, query, TransactionerMock{})

				_, err := svc.Variance(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTaxRuleResolver_Create(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "persists maps",
			fn: func(t *testing.T) {
				ctx := context.Background()
				var capturedTaxMaps []*TaxRuleTaxMap
				var capturedAccountMaps []*TaxRuleAccountMap
				positions := TaxRuleDAOMock{
					CreateWithMapsTxFunc: func(_ context.Context, _ *gorm.DB, position *TaxRule, taxMaps []*TaxRuleTaxMap, accountMaps []*TaxRuleAccountMap) (*TaxRule, error) {
						position.ID = 1
						for _, taxMap := range taxMaps {
							taxMap.TaxRuleID = position.ID
						}
						for _, accountMap := range accountMaps {
							accountMap.TaxRuleID = position.ID
						}
						capturedTaxMaps = taxMaps
						capturedAccountMaps = accountMaps
						return position, nil
					},
				}
				resolver := NewTaxRuleResolver(positions, TaxRuleTaxMapDAOMock{}, TaxRuleAccountMapDAOMock{}, TransactionerMock{})

				position, err := resolver.Create(ctx, CreateTaxRuleRequest{
					OrganizationID: 10,
					Name:           "ID",
					CountryCode:    "ID",
					TaxMaps:        []*TaxRuleTaxMap{{SrcTaxID: 9, DestTaxID: helper.Ptr(uint64(10))}},
					AccountMaps:    []*TaxRuleAccountMap{{SrcAccountID: 4100, DestAccountID: 4101}},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if position.Name == nil || *position.Name != "ID" {
					t.Errorf("name = %v, want ID", position.Name)
				}
				if len(capturedTaxMaps) != 1 || capturedTaxMaps[0].TaxRuleID != 1 {
					t.Errorf("tax maps = %+v, want map on position 1", capturedTaxMaps)
				}
				if len(capturedAccountMaps) != 1 || capturedAccountMaps[0].TaxRuleID != 1 {
					t.Errorf("account maps = %+v, want map on position 1", capturedAccountMaps)
				}
			},
		},
		{
			name: "propagates error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				positions := TaxRuleDAOMock{
					CreateWithMapsTxFunc: func(_ context.Context, _ *gorm.DB, _ *TaxRule, _ []*TaxRuleTaxMap, _ []*TaxRuleAccountMap) (*TaxRule, error) {
						return nil, errors.New("insert failed")
					},
				}
				resolver := NewTaxRuleResolver(positions, TaxRuleTaxMapDAOMock{}, TaxRuleAccountMapDAOMock{}, TransactionerMock{})

				_, err := resolver.Create(ctx, CreateTaxRuleRequest{OrganizationID: 10})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTaxRuleResolver_Resolve(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "remaps tax and account",
			fn: func(t *testing.T) {
				ctx := context.Background()
				position := &TaxRule{Base: model.Base{ID: 1}}
				positions := TaxRuleDAOMock{
					CRUDMock: dao.CRUDMock[TaxRule]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) { return position, nil },
					},
				}
				taxMaps := TaxRuleTaxMapDAOMock{
					ListByPositionFunc: func(_ context.Context, _ uint64) ([]*TaxRuleTaxMap, error) {
						return []*TaxRuleTaxMap{{SrcTaxID: 9, DestTaxID: helper.Ptr(uint64(10))}}, nil
					},
				}
				accountMaps := TaxRuleAccountMapDAOMock{
					ListByPositionFunc: func(_ context.Context, _ uint64) ([]*TaxRuleAccountMap, error) {
						return []*TaxRuleAccountMap{{SrcAccountID: 4100, DestAccountID: 4101}}, nil
					},
				}
				resolver := NewTaxRuleResolver(positions, taxMaps, accountMaps, TransactionerMock{})

				result, err := resolver.Resolve(ctx, 1, helper.Ptr(uint64(9)), 4100)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result.TaxAccount == nil || *result.TaxAccount != 10 {
					t.Errorf("tax account = %v, want 10", result.TaxAccount)
				}
				if result.Account == nil || *result.Account != 4101 {
					t.Errorf("account = %v, want 4101", result.Account)
				}
			},
		},
		{
			name: "rejects missing",
			fn: func(t *testing.T) {
				ctx := context.Background()
				positions := TaxRuleDAOMock{
					CRUDMock: dao.CRUDMock[TaxRule]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) { return nil, nil },
					},
				}
				resolver := NewTaxRuleResolver(positions, TaxRuleTaxMapDAOMock{}, TaxRuleAccountMapDAOMock{}, TransactionerMock{})

				_, err := resolver.Resolve(ctx, 1, helper.Ptr(uint64(9)), 4100)
				if helper.AssertError(t, err, true, ErrTaxRuleNotFound) {
					return
				}
			},
		},
		{
			name: "no tax mapping keeps source account",
			fn: func(t *testing.T) {
				ctx := context.Background()
				position := &TaxRule{Base: model.Base{ID: 1}}
				positions := TaxRuleDAOMock{
					CRUDMock: dao.CRUDMock[TaxRule]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) { return position, nil },
					},
				}
				accountMaps := TaxRuleAccountMapDAOMock{
					ListByPositionFunc: func(_ context.Context, _ uint64) ([]*TaxRuleAccountMap, error) {
						return []*TaxRuleAccountMap{{SrcAccountID: 4100, DestAccountID: 4101}}, nil
					},
				}
				resolver := NewTaxRuleResolver(positions, TaxRuleTaxMapDAOMock{}, accountMaps, TransactionerMock{})

				result, err := resolver.Resolve(ctx, 1, nil, 4100)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result.TaxAccount != nil {
					t.Errorf("tax account = %v, want nil", result.TaxAccount)
				}
				if result.Account == nil || *result.Account != 4101 {
					t.Errorf("account = %v, want 4101", result.Account)
				}
			},
		},
		{
			name: "propagates account map error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				position := &TaxRule{Base: model.Base{ID: 1}}
				positions := TaxRuleDAOMock{
					CRUDMock: dao.CRUDMock[TaxRule]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) { return position, nil },
					},
				}
				accountMaps := TaxRuleAccountMapDAOMock{
					ListByPositionFunc: func(_ context.Context, _ uint64) ([]*TaxRuleAccountMap, error) {
						return nil, errors.New("db down")
					},
				}
				resolver := NewTaxRuleResolver(positions, TaxRuleTaxMapDAOMock{}, accountMaps, TransactionerMock{})

				_, err := resolver.Resolve(ctx, 1, helper.Ptr(uint64(9)), 4100)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates tax map error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				position := &TaxRule{Base: model.Base{ID: 1}}
				positions := TaxRuleDAOMock{
					CRUDMock: dao.CRUDMock[TaxRule]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) { return position, nil },
					},
				}
				taxMaps := TaxRuleTaxMapDAOMock{
					ListByPositionFunc: func(_ context.Context, _ uint64) ([]*TaxRuleTaxMap, error) {
						return nil, errors.New("db down")
					},
				}
				resolver := NewTaxRuleResolver(positions, taxMaps, TaxRuleAccountMapDAOMock{}, TransactionerMock{})

				_, err := resolver.Resolve(ctx, 1, helper.Ptr(uint64(9)), 4100)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestWithholdingService_Withhold(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "posts three-way entry",
			fn: func(t *testing.T) {
				ctx := context.Background()
				date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
				withholdings := WithholdingTaxDAOMock{
					CRUDMock: dao.CRUDMock[WithholdingTax]{
						FindFunc: func(_ context.Context, _ uint64) (*WithholdingTax, error) {
							return &WithholdingTax{Base: model.Base{ID: 3}, RatePct: 2, AccountID: helper.Ptr(uint64(2200)), Scope: WithholdingScopePurchase}, nil
						},
					},
				}
				var capturedLines []PostingLine
				poster := PosterMock{
					PostFunc: func(_ context.Context, request PostRequest) (*JournalEntry, error) {
						capturedLines = request.Lines
						return &JournalEntry{Base: model.Base{ID: 60}}, nil
					},
				}
				svc := NewWithholdingService(withholdings, poster, TransactionerMock{})

				entry, err := svc.Withhold(ctx, WithholdRequest{
					OrganizationID: 10,
					JournalID:      20,
					Amount:         1000,
					Date:           date,
					Scope:          WithholdingScopePurchase,
					WHTID:          3,
					BankAccountID:  1100,
					PayableID:      2201,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if entry.ID != 60 {
					t.Errorf("entry id = %d, want 60", entry.ID)
				}
				if len(capturedLines) != 3 {
					t.Fatalf("lines = %d, want 3", len(capturedLines))
				}
				if capturedLines[0].Debit.Float64() != 1000 {
					t.Errorf("payable debit = %v, want 1000", capturedLines[0].Debit)
				}
				if capturedLines[1].Credit.Float64() != 20 {
					t.Errorf("wht credit = %v, want 20", capturedLines[1].Credit)
				}
				if capturedLines[2].Credit.Float64() != 980 {
					t.Errorf("bank credit = %v, want 980", capturedLines[2].Credit)
				}
			},
		},
		{
			name: "rejects wrong scope",
			fn: func(t *testing.T) {
				ctx := context.Background()
				withholdings := WithholdingTaxDAOMock{
					CRUDMock: dao.CRUDMock[WithholdingTax]{
						FindFunc: func(_ context.Context, _ uint64) (*WithholdingTax, error) {
							return &WithholdingTax{Base: model.Base{ID: 3}, RatePct: 2, AccountID: helper.Ptr(uint64(2200)), Scope: WithholdingScopeSale}, nil
						},
					},
				}
				svc := NewWithholdingService(withholdings, PosterMock{}, TransactionerMock{})

				_, err := svc.Withhold(ctx, WithholdRequest{OrganizationID: 10, WHTID: 3, Scope: WithholdingScopePurchase})
				if helper.AssertError(t, err, true, ErrWithholdingScope) {
					return
				}
			},
		},
		{
			name: "rejects missing",
			fn: func(t *testing.T) {
				ctx := context.Background()
				withholdings := WithholdingTaxDAOMock{
					CRUDMock: dao.CRUDMock[WithholdingTax]{
						FindFunc: func(_ context.Context, _ uint64) (*WithholdingTax, error) { return nil, nil },
					},
				}
				svc := NewWithholdingService(withholdings, PosterMock{}, TransactionerMock{})

				_, err := svc.Withhold(ctx, WithholdRequest{OrganizationID: 10, WHTID: 3, Scope: WithholdingScopePurchase})
				if helper.AssertError(t, err, true, ErrWithholdingNotFound) {
					return
				}
			},
		},
		{
			name: "rejects without account",
			fn: func(t *testing.T) {
				ctx := context.Background()
				withholdings := WithholdingTaxDAOMock{
					CRUDMock: dao.CRUDMock[WithholdingTax]{
						FindFunc: func(_ context.Context, _ uint64) (*WithholdingTax, error) {
							return &WithholdingTax{Base: model.Base{ID: 3}, RatePct: 2, Scope: WithholdingScopePurchase}, nil
						},
					},
				}
				svc := NewWithholdingService(withholdings, PosterMock{}, TransactionerMock{})

				_, err := svc.Withhold(ctx, WithholdRequest{OrganizationID: 10, WHTID: 3, Scope: WithholdingScopePurchase})
				if helper.AssertError(t, err, true, ErrWithholdingNoAccount) {
					return
				}
			},
		},
		{
			name: "propagates find error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				withholdings := WithholdingTaxDAOMock{
					CRUDMock: dao.CRUDMock[WithholdingTax]{
						FindFunc: func(_ context.Context, _ uint64) (*WithholdingTax, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewWithholdingService(withholdings, PosterMock{}, TransactionerMock{})

				_, err := svc.Withhold(ctx, WithholdRequest{OrganizationID: 10, WHTID: 3, Scope: WithholdingScopePurchase})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTaxReturnService_ExportFiling(t *testing.T) {
	filedAt := time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)
	filed := &TaxReturn{
		Base:           model.Base{ID: 3},
		OrganizationID: helper.Ptr(uint64(10)),
		PeriodID:       1,
		Type:           "vat",
		OutputTax:      220,
		InputTax:       100,
		NetPayable:     120,
		State:          TaxReturnStateFiled,
		FiledAt:        &filedAt,
	}
	period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, Name: "Jan 2026"}

	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return filed, nil },
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
		},
	}
	svc := NewTaxReturnService(returns, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

	content, err := svc.ExportFiling(context.Background(), 3, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "tax_return_id,organization_id,period_id,period,type,state,filed_at,output_tax,input_tax,net_payable\n" +
		"3,10,1,Jan 2026,vat,filed,2026-02-05,220.0000,100.0000,120.0000\n"
	if content != want {
		t.Errorf("csv = %q, want %q", content, want)
	}
}

func TestTaxReturnService_ExportFilingRejects(t *testing.T) {
	tests := []struct {
		name    string
		return_ *TaxReturn
		orgID   uint64
		wantErr error
	}{
		{
			name:    "draft cannot export",
			return_: &TaxReturn{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(10)), PeriodID: 1, State: TaxReturnStateDraft},
			orgID:   10,
			wantErr: ErrTaxReturnNotFiled,
		},
		{
			name:    "other organization not found",
			return_: &TaxReturn{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(99)), PeriodID: 1, State: TaxReturnStateFiled},
			orgID:   10,
			wantErr: ErrTaxReturnNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returns := TaxReturnDAOMock{
				CRUDMock: dao.CRUDMock[TaxReturn]{
					FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return tt.return_, nil },
				},
			}
			svc := NewTaxReturnService(returns, TaxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})
			if _, err := svc.ExportFiling(context.Background(), 3, tt.orgID); err != tt.wantErr {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxReturnService_Create(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "rejects missing period",
			fn: func(t *testing.T) {
				ctx := context.Background()
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return nil, nil },
					},
				}
				svc := NewTaxReturnService(TaxReturnDAOMock{}, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, ErrPeriodNotFound) {
					return
				}
			},
		},
		{
			name: "rejects period of another organization",
			fn: func(t *testing.T) {
				ctx := context.Background()
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
				period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 99, DateStart: &start, DateEnd: &end}
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
					},
				}
				svc := NewTaxReturnService(TaxReturnDAOMock{}, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, ErrPeriodNotFound) {
					return
				}
			},
		},
		{
			name: "propagates find by period error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
				period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end}
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
					},
				}
				returns := TaxReturnDAOMock{
					FindByPeriodFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewTaxReturnService(returns, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates output tax error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
				period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end}
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
					},
				}
				invoiceTaxes := InvoiceTaxDAOMock{
					SumTaxByPeriodFunc: func(_ context.Context, _ uint64, _ string, _, _ time.Time) (float64, error) {
						return 0, errors.New("db down")
					},
				}
				svc := NewTaxReturnService(TaxReturnDAOMock{}, periods, invoiceTaxes, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates input tax error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
				period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end}
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
					},
				}
				perType := map[string]error{InvoiceTypeCustomerInvoice: nil, InvoiceTypeSupplierBill: errors.New("db down")}
				invoiceTaxes := InvoiceTaxDAOMock{
					SumTaxByPeriodFunc: func(_ context.Context, _ uint64, invoiceType string, _, _ time.Time) (float64, error) {
						return 0, perType[invoiceType]
					},
				}
				svc := NewTaxReturnService(TaxReturnDAOMock{}, periods, invoiceTaxes, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			name: "propagates period find error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				periods := taxPeriodDAOMock{
					CRUDMock: dao.CRUDMock[TaxPeriod]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewTaxReturnService(TaxReturnDAOMock{}, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTaxReturnService_File(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "rejects missing",
			fn: func(t *testing.T) {
				ctx := context.Background()
				returns := TaxReturnDAOMock{
					CRUDMock: dao.CRUDMock[TaxReturn]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return nil, nil },
					},
				}
				svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.File(ctx, 1)
				if helper.AssertError(t, err, true, ErrTaxReturnNotFound) {
					return
				}
			},
		},
		{
			name: "rejects not draft",
			fn: func(t *testing.T) {
				ctx := context.Background()
				taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateFiled}
				returns := TaxReturnDAOMock{
					CRUDMock: dao.CRUDMock[TaxReturn]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
					},
				}
				svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.File(ctx, 1)
				if helper.AssertError(t, err, true, ErrTaxReturnNotDraft) {
					return
				}
			},
		},
		{
			name: "propagates find error",
			fn: func(t *testing.T) {
				ctx := context.Background()
				returns := TaxReturnDAOMock{
					CRUDMock: dao.CRUDMock[TaxReturn]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.File(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTaxReturnService_Pay(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "not found",
			fn: func(t *testing.T) {
				ctx := context.Background()
				returns := TaxReturnDAOMock{
					CRUDMock: dao.CRUDMock[TaxReturn]{
						FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return nil, nil },
					},
				}
				svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

				_, err := svc.Pay(ctx, 1)
				if helper.AssertError(t, err, true, ErrTaxReturnNotFound) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
