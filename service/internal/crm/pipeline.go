package crm

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type PipelineService struct {
	leads  ProspectDAO
	stages dao.CRUD[reference.PipelineStage]
}

func NewPipelineService(leads ProspectDAO, stages dao.CRUD[reference.PipelineStage]) PipelineService {
	return PipelineService{leads: leads, stages: stages}
}

type PipelineStage struct {
	StageID          uint64  `json:"stage_id"`
	StageName        string  `json:"stage_name"`
	Probability      float64 `json:"probability"`
	OpportunityCount int64   `json:"opportunity_count"`
	ExpectedRevenue  float64 `json:"expected_revenue"`
	WeightedRevenue  float64 `json:"weighted_revenue"`
}

type PipelineForecast struct {
	TotalExpectedRevenue float64         `json:"total_expected_revenue"`
	WeightedPipeline     float64         `json:"weighted_pipeline"`
	WinRate              float64         `json:"win_rate"`
	Stages               []PipelineStage `json:"stages"`
}

func (s PipelineService) Forecast(ctx context.Context, organizationID *uint64) (PipelineForecast, error) {
	parsedQuery := &query.Query{
		Sorts: []query.Sort{{Field: "sequence", Direction: query.Ascending}},
	}
	if organizationID != nil {
		parsedQuery.Filters = []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: *organizationID}}
	}
	page, err := s.stages.List(ctx, parsedQuery)
	if err != nil {
		return PipelineForecast{}, err
	}

	open, err := s.leads.ListOpen(ctx, organizationID)
	if err != nil {
		return PipelineForecast{}, err
	}
	won, err := s.leads.CountWon(ctx, organizationID)
	if err != nil {
		return PipelineForecast{}, err
	}
	lost, err := s.leads.CountLost(ctx, organizationID)
	if err != nil {
		return PipelineForecast{}, err
	}

	stages := make([]PipelineStage, 0, len(page.Items))
	index := make(map[uint64]int, len(page.Items))
	for _, stage := range page.Items {
		index[stage.ID] = len(stages)
		stages = append(stages, PipelineStage{
			StageID:     stage.ID,
			StageName:   stage.Name,
			Probability: stage.Probability,
		})
	}

	totalExpected := amount.Zero()
	weighted := amount.Zero()
	for _, prospect := range open {
		expected := amount.FromFloat64(prospect.ExpectedRevenue)
		weightedRow, err := weightedAmount(prospect.ExpectedRevenue, prospect.Probability)
		if err != nil {
			return PipelineForecast{}, err
		}
		totalExpected = totalExpected.Add(expected)
		weighted = weighted.Add(weightedRow)

		if prospect.StageID != nil {
			if idx, ok := index[*prospect.StageID]; ok {
				stages[idx].OpportunityCount++
				stages[idx].ExpectedRevenue += expected.Float64()
				stages[idx].WeightedRevenue += weightedRow.Float64()
			}
		}
	}

	return PipelineForecast{
		TotalExpectedRevenue: totalExpected.Round(4).Float64(),
		WeightedPipeline:     weighted.Round(4).Float64(),
		WinRate:              winRate(won, lost),
		Stages:               stages,
	}, nil
}

func weightedAmount(expected, probability float64) (amount.Amount, error) {
	probFactor, err := amount.FromFloat64(probability).Div(amount.FromInt64(100))
	if err != nil {
		return amount.Amount{}, err
	}
	return amount.FromFloat64(expected).Mul(probFactor).Round(4), nil
}

func winRate(won, lost int64) float64 {
	denominator := won + lost
	if denominator == 0 {
		return 0
	}
	rate, err := amount.FromInt64(won).Div(amount.FromInt64(denominator))
	if err != nil {
		return 0
	}
	return rate.Mul(amount.FromInt64(100)).Round(2).Float64()
}
