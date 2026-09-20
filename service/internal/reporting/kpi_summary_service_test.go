package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

func TestKpiSummaryService_RefreshPeriod_MergesOpeningAndActivity(t *testing.T) {
	ctx := context.Background()
	var replaced []KpiAccountSummary
	svc := NewKpiSummaryService(
		KpiSummaryDAOMock{
			OpeningBalancesFn: func(_ context.Context, _ uint64, _ time.Time) ([]AccountPeriodBalance, error) {
				return []AccountPeriodBalance{
					{AccountID: 11, Debit: 100},
					{AccountID: 12, Credit: 100},
				}, nil
			},
			PeriodActivityFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountPeriodBalance, error) {
				return []AccountPeriodBalance{
					{AccountID: 11, Debit: 50},
					{AccountID: 13, Credit: 60},
				}, nil
			},
			ReplacePeriodFn: func(_ context.Context, _, _ uint64, rows []KpiAccountSummary) error {
				replaced = rows
				return nil
			},
		},
		TaxPeriodFinderMock{},
	)

	processed, err := svc.RefreshPeriod(ctx, 1, 7)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if processed != 3 {
		t.Errorf("processed = %d, want 3", processed)
	}
	if len(replaced) != 3 {
		t.Fatalf("replaced rows = %d, want 3", len(replaced))
	}
	byAccount := map[uint64]KpiAccountSummary{}
	for _, row := range replaced {
		byAccount[row.AccountID] = row
	}
	if got := byAccount[11]; got.OpeningDebit != 100 || got.PeriodDebit != 50 {
		t.Errorf("account 11 = %+v, want opening 100 period 50", got)
	}
	if got := byAccount[12]; got.OpeningCredit != 100 || got.PeriodCredit != 0 {
		t.Errorf("account 12 = %+v, want opening credit 100", got)
	}
	if got := byAccount[13]; got.PeriodCredit != 60 {
		t.Errorf("account 13 = %+v, want period credit 60", got)
	}
}

func TestKpiSummaryService_RefreshPeriod_PeriodNotFound(t *testing.T) {
	svc := NewKpiSummaryService(KpiSummaryDAOMock{}, TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		},
	})

	if _, err := svc.RefreshPeriod(context.Background(), 1, 7); !errors.Is(err, ErrPeriodNotFound) {
		t.Errorf("err = %v, want ErrPeriodNotFound", err)
	}
}
