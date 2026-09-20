package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestTrialBalanceService_Generate(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{
				Base:           model.Base{ID: id},
				OrganizationID: 10,
				DateStart:      &start,
				DateEnd:        &end,
				State:          TaxPeriodStateOpen,
			}, nil
		}},
	}
	tbDAO := TrialBalanceDAOMock{
		GenerateByPeriodFunc: func(_ context.Context, orgID, periodID uint64) ([]TrialBalanceLine, error) {
			return []TrialBalanceLine{
				{AccountID: 1, AccountCode: "1100", AccountName: "Cash", AccountType: "cash", Debit: amount.FromFloat64(5000), Credit: amount.Zero(), Balance: amount.FromFloat64(5000)},
				{AccountID: 2, AccountCode: "4100", AccountName: "Revenue", AccountType: "income", Debit: amount.Zero(), Credit: amount.FromFloat64(3000), Balance: amount.FromFloat64(-3000)},
			}, nil
		},
	}

	svc := NewTrialBalanceService(tbDAO, periods)
	tb, err := svc.Generate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tb.Lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(tb.Lines))
	}
	if !tb.TotalDebit.Equal(amount.FromFloat64(5000)) {
		t.Errorf("total debit = %v, want 5000", tb.TotalDebit)
	}
	if !tb.TotalCredit.Equal(amount.FromFloat64(3000)) {
		t.Errorf("total credit = %v, want 3000", tb.TotalCredit)
	}
	if tb.PeriodID != 1 {
		t.Errorf("period id = %d, want 1", tb.PeriodID)
	}
}

func TestTrialBalanceService_Generate_RejectsWrongOrg(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{
				Base:           model.Base{ID: id},
				OrganizationID: 10,
				DateStart:      &start,
				DateEnd:        &end,
				State:          TaxPeriodStateOpen,
			}, nil
		}},
	}
	svc := NewTrialBalanceService(TrialBalanceDAOMock{}, periods)

	_, err := svc.Generate(ctx, 99, 1)
	if err != ErrPeriodNotFound {
		t.Errorf("err = %v, want ErrPeriodNotFound", err)
	}
}
