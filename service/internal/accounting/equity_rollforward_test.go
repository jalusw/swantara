package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestEquityService_Generate(t *testing.T) {
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
	equityDAO := EquityDAOMock{
		OpeningEquityBalanceFunc: func(_ context.Context, _ uint64, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{
				{AccountID: 100, AccountCode: "3100", AccountName: "Share Capital", AccountType: "equity", Amount: 50000},
			}, nil
		},
		EquityBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{
				{AccountID: 100, AccountCode: "3100", AccountName: "Share Capital", AccountType: "equity", Amount: 10000},
			}, nil
		},
		NetIncomeForPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) {
			return 2500, nil
		},
	}

	svc := NewEquityService(equityDAO, periods)
	eq, err := svc.Generate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eq.NetIncome != 2500 {
		t.Errorf("net income = %f, want 2500", eq.NetIncome)
	}
	if len(eq.OpeningBalance) != 1 {
		t.Errorf("opening balance lines = %d, want 1", len(eq.OpeningBalance))
	}
	if len(eq.ClosingBalance) != 1 {
		t.Errorf("closing balance lines = %d, want 1", len(eq.ClosingBalance))
	}
	if eq.ClosingBalance[0].Amount != 60000 {
		t.Errorf("closing amount = %f, want 60000", eq.ClosingBalance[0].Amount)
	}
}

func TestEquityService_RejectsWrongOrg(t *testing.T) {
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
	svc := NewEquityService(EquityDAOMock{}, periods)

	_, err := svc.Generate(ctx, 99, 1)
	if err != ErrPeriodNotFound {
		t.Errorf("err = %v, want ErrPeriodNotFound", err)
	}
}
