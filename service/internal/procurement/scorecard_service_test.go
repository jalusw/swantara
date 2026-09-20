package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func testScorecardService() (SupplierScorecardService, *SupplierScorecardDAOMock) {
	scorecards := &SupplierScorecardDAOMock{}
	svc := NewSupplierScorecardService(scorecards)
	return svc, scorecards
}

func TestSupplierScorecardService_ComputeScores_AllPassing(t *testing.T) {
	svc, _ := testScorecardService()
	quality, delivery, price, overall := svc.ComputeScores(100, 95, 2, 5)
	if quality != 98 {
		t.Errorf("expected quality 98, got %f", quality)
	}
	if delivery != 95 {
		t.Errorf("expected delivery 95, got %f", delivery)
	}
	if price != 75 {
		t.Errorf("expected price 75, got %f", price)
	}
	expectedOverall := 98*0.4 + 95*0.3 + 75*0.3
	if overall != expectedOverall {
		t.Errorf("expected overall %f, got %f", expectedOverall, overall)
	}
}

func TestSupplierScorecardService_ComputeScores_ZeroOrders(t *testing.T) {
	svc, _ := testScorecardService()
	quality, delivery, price, overall := svc.ComputeScores(0, 0, 0, 0)
	if quality != 0 || delivery != 0 || price != 0 || overall != 0 {
		t.Errorf("expected all zeros, got quality=%f delivery=%f price=%f overall=%f", quality, delivery, price, overall)
	}
}

func TestSupplierScorecardService_ComputeScores_AllFailed(t *testing.T) {
	svc, _ := testScorecardService()
	quality, delivery, price, overall := svc.ComputeScores(100, 0, 100, 30)
	if quality != 0 {
		t.Errorf("expected quality 0, got %f", quality)
	}
	if delivery != 0 {
		t.Errorf("expected delivery 0, got %f", delivery)
	}
	if price != 0 {
		t.Errorf("expected price 0, got %f", price)
	}
	if overall != 0 {
		t.Errorf("expected overall 0, got %f", overall)
	}
}

func TestSupplierScorecardService_ComputeScores_PerfectScore(t *testing.T) {
	svc, _ := testScorecardService()
	quality, delivery, price, overall := svc.ComputeScores(100, 100, 0, 0)
	if quality != 100 {
		t.Errorf("expected quality 100, got %f", quality)
	}
	if delivery != 100 {
		t.Errorf("expected delivery 100, got %f", delivery)
	}
	if price != 100 {
		t.Errorf("expected price 100, got %f", price)
	}
	if overall != 100 {
		t.Errorf("expected overall 100, got %f", overall)
	}
}

func TestSupplierScorecardService_UpdateScorecard_CreatesNew(t *testing.T) {
	ctx := context.Background()
	svc, scorecards := testScorecardService()
	scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
		return nil, nil
	}
	scorecards.CreateFunc = func(_ context.Context, sc *SupplierScorecard) (*SupplierScorecard, error) {
		sc.ID = 1
		return sc, nil
	}
	doneAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	err := svc.UpdateScorecard(ctx, nil, 10, doneAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplierScorecardService_UpdateScorecard_UpdatesExisting(t *testing.T) {
	ctx := context.Background()
	svc, scorecards := testScorecardService()
	scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
		return &SupplierScorecard{
			Base:             model.Base{ID: 1},
			SupplierID:       10,
			TotalOrders:      5,
			OnTimeDeliveries: 4,
			QualityFailures:  1,
		}, nil
	}
	scorecards.UpdateFunc = func(_ context.Context, sc *SupplierScorecard) (*SupplierScorecard, error) {
		if sc.TotalOrders != 6 {
			t.Errorf("expected TotalOrders 6, got %d", sc.TotalOrders)
		}
		if sc.OnTimeDeliveries != 5 {
			t.Errorf("expected OnTimeDeliveries 5, got %d", sc.OnTimeDeliveries)
		}
		return sc, nil
	}
	doneAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	err := svc.UpdateScorecard(ctx, nil, 10, doneAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplierScorecardService_RecordQualityFailure(t *testing.T) {
	ctx := context.Background()
	svc, scorecards := testScorecardService()
	scorecards.FindByVendorAndPeriodFunc = func(_ context.Context, _ uint64, _, _ *time.Time) (*SupplierScorecard, error) {
		return &SupplierScorecard{
			Base:             model.Base{ID: 1},
			SupplierID:       10,
			TotalOrders:      5,
			OnTimeDeliveries: 4,
			QualityFailures:  1,
		}, nil
	}
	scorecards.UpdateFunc = func(_ context.Context, sc *SupplierScorecard) (*SupplierScorecard, error) {
		if sc.TotalOrders != 6 {
			t.Errorf("expected TotalOrders 6, got %d", sc.TotalOrders)
		}
		if sc.QualityFailures != 2 {
			t.Errorf("expected QualityFailures 2, got %d", sc.QualityFailures)
		}
		return sc, nil
	}
	doneAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	err := svc.RecordQualityFailure(ctx, nil, 10, doneAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplierScorecardService_ListBySupplier_Success(t *testing.T) {
	ctx := context.Background()
	svc, scorecards := testScorecardService()
	scorecards.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[SupplierScorecard], error) {
		return &query.Page[SupplierScorecard]{
			Items: []*SupplierScorecard{
				{Base: model.Base{ID: 1}, SupplierID: 10, OverallScore: 85},
				{Base: model.Base{ID: 2}, SupplierID: 10, OverallScore: 90},
			},
			Count: 2,
		}, nil
	}
	result, err := svc.ListBySupplier(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 scorecards, got %d", len(result))
	}
}
