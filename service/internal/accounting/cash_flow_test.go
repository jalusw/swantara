package accounting

import (
	"context"
	"testing"
	"time"
)

func TestCashFlowService_Generate(t *testing.T) {
	ctx := context.Background()
	dateStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	dateEnd := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	dao := CashFlowDAOMock{
		GetOpeningCashFunc: func(_ context.Context, _ uint64, _ time.Time) (float64, error) {
			return 10000, nil
		},
		GenerateByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowLine, error) {
			return []CashFlowLine{
				{AccountID: 1, AccountCode: "1100", Name: "Cash receipt", Amount: 5000, CashFlowSection: CashFlowOperating, OriginType: OriginTypeSaleOrder},
				{AccountID: 2, AccountCode: "1500", Name: "Equipment purchase", Amount: -3000, CashFlowSection: CashFlowInvesting, OriginType: OriginTypeInboundCost},
			}, nil
		},
	}

	svc := NewCashFlowService(dao)
	cf, err := svc.Generate(ctx, 10, dateStart, dateEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.OpeningCash != 10000 {
		t.Errorf("opening cash = %f, want 10000", cf.OpeningCash)
	}
	if len(cf.Operating) != 1 {
		t.Errorf("operating lines = %d, want 1", len(cf.Operating))
	}
	if len(cf.Investing) != 1 {
		t.Errorf("investing lines = %d, want 1", len(cf.Investing))
	}
	if cf.NetChange != 2000 {
		t.Errorf("net change = %f, want 2000", cf.NetChange)
	}
	if cf.ClosingCash != 12000 {
		t.Errorf("closing cash = %f, want 12000", cf.ClosingCash)
	}
}

func TestCashFlowService_Generate_Empty(t *testing.T) {
	ctx := context.Background()
	dateStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	dateEnd := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	dao := CashFlowDAOMock{
		GetOpeningCashFunc: func(_ context.Context, _ uint64, _ time.Time) (float64, error) {
			return 5000, nil
		},
		GenerateByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowLine, error) {
			return []CashFlowLine{}, nil
		},
	}

	svc := NewCashFlowService(dao)
	cf, err := svc.Generate(ctx, 10, dateStart, dateEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.OpeningCash != 5000 {
		t.Errorf("opening cash = %f, want 5000", cf.OpeningCash)
	}
	if cf.ClosingCash != 5000 {
		t.Errorf("closing cash = %f, want 5000", cf.ClosingCash)
	}
	if cf.NetChange != 0 {
		t.Errorf("net change = %f, want 0", cf.NetChange)
	}
}

func TestClassifyCashFlowByOrigin(t *testing.T) {
	tests := []struct {
		origin string
		want   string
	}{
		{OriginTypeSaleOrder, CashFlowOperating},
		{OriginTypePurchaseOrder, CashFlowOperating},
		{OriginTypeInboundCost, CashFlowInvesting},
		{"", CashFlowOperating},
	}
	for _, tt := range tests {
		got := classifyCashFlowByOrigin(tt.origin, "")
		if got != tt.want {
			t.Errorf("classify %q = %q, want %q", tt.origin, got, tt.want)
		}
	}
}
