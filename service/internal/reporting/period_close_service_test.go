package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestPeriodCloseService_Close_RunsEveryStepThenLocks(t *testing.T) {
	ctx := context.Background()
	closer := &PeriodCloserMock{}
	depreciationPosted := 0
	depreciation := &DepreciationPosterMock{PostFn: func(_ context.Context, _ asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error) {
		if depreciationPosted == 1 {
			return nil, asset.ErrAssetNothingToPost
		}
		depreciationPosted++
		return &asset.AssetDepreciationLine{Posted: true}, nil
	}}
	fx := NewFxRevaluationService(ReportDAOMock{}, FxRevaluationDAOMock{}, FxRevaluationLineDAOMock{}, &PosterMock{}, ConfigSourceMock{}, RateResolverMock{}, OrgReaderMock{})
	accruals := NewAccrualService(AccrualDAOMock{}, AccrualLineDAOMock{}, &PosterMock{}, ConfigSourceMock{})
	deferrals := DeferralRecognizerMock{
		RecognizeDueFn: func(_ context.Context, _ *uint64, _ time.Time) (int, error) { return 3, nil },
	}
	summaries := KpiSummarizerMock{
		RefreshPeriodFn: func(_ context.Context, organizationID, periodID uint64) (int, error) {
			return 5, nil
		},
	}
	svc := NewPeriodCloseService(
		closer,
		TaxPeriodFinderMock{},
		ConfigSourceMock{},
		AssetListReaderMock{ListFn: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
			return &query.Page[asset.FixedAsset]{Items: []*asset.FixedAsset{{State: asset.AssetStateRunning}, {State: asset.AssetStateDisposed}}}, nil
		}},
		depreciation,
		fx,
		accruals,
		deferrals,
		summaries,
	)

	result, err := svc.Close(ctx, 1, 7)
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if depreciationPosted != 1 {
		t.Errorf("depreciation postings = %d, want 1 (only running asset)", depreciationPosted)
	}
	if result.DepreciationLines != 1 || result.DeferralsRecognized != 3 {
		t.Errorf("result = %+v, want 1 depreciation, 3 deferrals", result)
	}
	if result.SummarizedAccounts != 5 {
		t.Errorf("summarized accounts = %d, want 5", result.SummarizedAccounts)
	}
	if closer.closeCount != 1 || closer.lockCount != 1 {
		t.Errorf("close/lock calls = %d/%d, want 1/1", closer.closeCount, closer.lockCount)
	}
}

func TestPeriodCloseService_Close_RejectsLockedPeriod(t *testing.T) {
	ctx := context.Background()
	locked := accounting.TaxPeriodStateLocked
	svc := NewPeriodCloseService(
		&PeriodCloserMock{},
		TaxPeriodFinderMock{FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{State: locked}, nil
		}},
		ConfigSourceMock{},
		AssetListReaderMock{},
		&DepreciationPosterMock{},
		NewFxRevaluationService(ReportDAOMock{}, FxRevaluationDAOMock{}, FxRevaluationLineDAOMock{}, &PosterMock{}, ConfigSourceMock{}, RateResolverMock{}, OrgReaderMock{}),
		NewAccrualService(AccrualDAOMock{}, AccrualLineDAOMock{}, &PosterMock{}, ConfigSourceMock{}),
		DeferralRecognizerMock{},
		KpiSummarizerMock{},
	)

	_, err := svc.Close(ctx, 1, 7)
	if !errors.Is(err, ErrPeriodLocked) {
		t.Errorf("err = %v, want ErrPeriodLocked", err)
	}
}

func TestPeriodCloseService_Close_AbortsBeforeLockWhenDepreciationFails(t *testing.T) {
	ctx := context.Background()
	closer := &PeriodCloserMock{}
	svc := NewPeriodCloseService(
		closer,
		TaxPeriodFinderMock{},
		ConfigSourceMock{},
		AssetListReaderMock{ListFn: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
			return &query.Page[asset.FixedAsset]{Items: []*asset.FixedAsset{{State: asset.AssetStateRunning}}}, nil
		}},
		&DepreciationPosterMock{PostFn: func(_ context.Context, _ asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error) {
			return nil, errors.New("depreciation broke")
		}},
		NewFxRevaluationService(ReportDAOMock{}, FxRevaluationDAOMock{}, FxRevaluationLineDAOMock{}, &PosterMock{}, ConfigSourceMock{}, RateResolverMock{}, OrgReaderMock{}),
		NewAccrualService(AccrualDAOMock{}, AccrualLineDAOMock{}, &PosterMock{}, ConfigSourceMock{}),
		DeferralRecognizerMock{},
		KpiSummarizerMock{},
	)

	if _, err := svc.Close(ctx, 1, 7); err == nil {
		t.Fatal("expected close to fail on depreciation error")
	}
	if closer.closeCount != 0 || closer.lockCount != 0 {
		t.Errorf("close/lock calls = %d/%d, want period left untouched", closer.closeCount, closer.lockCount)
	}
}
