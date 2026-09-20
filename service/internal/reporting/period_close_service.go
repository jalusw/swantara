package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type PeriodCloser interface {
	Close(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error)
	Lock(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error)
}

type AssetListReader interface {
	List(ctx context.Context, q *query.Query) (*query.Page[asset.FixedAsset], error)
}

type DepreciationPoster interface {
	PostDepreciation(ctx context.Context, request asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error)
}

type DeferralRecognizer interface {
	RecognizeDue(ctx context.Context, organizationID *uint64, asOf time.Time) (int, error)
}

type KpiSummarizer interface {
	RefreshPeriod(ctx context.Context, organizationID, periodID uint64) (int, error)
}

type PeriodCloseService struct {
	periods      PeriodCloser
	finder       TaxPeriodFinder
	config       ConfigSource
	assets       AssetListReader
	depreciation DepreciationPoster
	fx           FxRevaluationService
	accruals     AccrualService
	deferrals    DeferralRecognizer
	summaries    KpiSummarizer
}

func NewPeriodCloseService(
	periods PeriodCloser,
	finder TaxPeriodFinder,
	config ConfigSource,
	assets AssetListReader,
	depreciation DepreciationPoster,
	fx FxRevaluationService,
	accruals AccrualService,
	deferrals DeferralRecognizer,
	summaries KpiSummarizer,
) PeriodCloseService {
	return PeriodCloseService{
		periods:      periods,
		finder:       finder,
		config:       config,
		assets:       assets,
		depreciation: depreciation,
		fx:           fx,
		accruals:     accruals,
		deferrals:    deferrals,
		summaries:    summaries,
	}
}

type CloseResult struct {
	PeriodID            uint64            `json:"period_id"`
	Date                time.Time         `json:"date"`
	DepreciationLines   int               `json:"depreciation_lines"`
	FxRevaluation       RevaluationResult `json:"fx_revaluation"`
	AccrualsReversed    int               `json:"accruals_reversed"`
	DeferralsRecognized int               `json:"deferrals_recognized"`
	SummarizedAccounts  int               `json:"summarized_accounts"`
}

func (s PeriodCloseService) Close(ctx context.Context, organizationID, periodID uint64) (CloseResult, error) {
	period, err := s.finder.Find(ctx, periodID)
	if err != nil {
		return CloseResult{}, err
	}
	if period == nil {
		return CloseResult{}, ErrPeriodNotFound
	}
	if period.State == accounting.TaxPeriodStateLocked {
		return CloseResult{}, ErrPeriodLocked
	}
	if period.DateEnd == nil {
		return CloseResult{}, ErrPeriodNotFound
	}
	end := *period.DateEnd

	journalID, err := s.config.JournalID(ctx, organizationID)
	if err != nil {
		return CloseResult{}, err
	}

	assetPage, err := s.assets.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return CloseResult{}, err
	}
	depreciation := 0
	for _, a := range assetPage.Items {
		if a.State != asset.AssetStateRunning {
			continue
		}
		for {
			_, err := s.depreciation.PostDepreciation(ctx, asset.PostDepreciationRequest{
				AssetID:   a.ID,
				JournalID: journalID,
				Date:      end,
			})
			if errors.Is(err, asset.ErrAssetNothingToPost) {
				break
			}
			if err != nil {
				return CloseResult{}, err
			}
			depreciation++
		}
	}

	fxResult, err := s.fx.Revalue(ctx, organizationID, periodID, end)
	if err != nil {
		return CloseResult{}, err
	}

	accrualsReversed, err := s.accruals.ReverseDue(ctx, organizationID, end)
	if err != nil {
		return CloseResult{}, err
	}

	deferralsRecognized, err := s.deferrals.RecognizeDue(ctx, &organizationID, end)
	if err != nil {
		return CloseResult{}, err
	}

	summarized, err := s.summaries.RefreshPeriod(ctx, organizationID, periodID)
	if err != nil {
		return CloseResult{}, err
	}

	if _, err := s.periods.Close(ctx, periodID); err != nil {
		return CloseResult{}, err
	}
	if _, err := s.periods.Lock(ctx, periodID); err != nil {
		return CloseResult{}, err
	}

	return CloseResult{
		PeriodID:            periodID,
		Date:                end,
		DepreciationLines:   depreciation,
		FxRevaluation:       fxResult,
		AccrualsReversed:    accrualsReversed,
		DeferralsRecognized: deferralsRecognized,
		SummarizedAccounts:  summarized,
	}, nil
}
