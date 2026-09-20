package crm

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestPipelineService_Forecast_WeightedPipelineAndWinRate(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{
				{Base: model.Base{ID: 1}, Name: "New", Sequence: 10, Probability: 10},
				{Base: model.Base{ID: 2}, Name: "Qualified", Sequence: 20, Probability: 35},
			}}, nil
		},
	}
	leads := ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*Prospect, error) {
			return []*Prospect{
				{StageID: &[]uint64{1}[0], ExpectedRevenue: 1000, Probability: 10},
				{StageID: &[]uint64{2}[0], ExpectedRevenue: 2000, Probability: 35},
			}, nil
		},
		CountWonFunc:  func(_ context.Context, _ *uint64) (int64, error) { return 2, nil },
		CountLostFunc: func(_ context.Context, _ *uint64) (int64, error) { return 1, nil },
	}
	svc := NewPipelineService(leads, stages)

	forecast, err := svc.Forecast(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forecast.TotalExpectedRevenue != 3000 {
		t.Errorf("total_expected_revenue = %v, want 3000", forecast.TotalExpectedRevenue)
	}
	if forecast.WeightedPipeline != 800 {
		t.Errorf("weighted_pipeline = %v, want 800", forecast.WeightedPipeline)
	}
	if forecast.WinRate != 66.67 {
		t.Errorf("win_rate = %v, want 66.67", forecast.WinRate)
	}
}

func TestPipelineService_Forecast_EmptyPipelineHasZeroWinRate(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{
				{Base: model.Base{ID: 1}, Name: "New", Sequence: 10, Probability: 10},
			}}, nil
		},
	}
	svc := NewPipelineService(ProspectDAOMock{}, stages)

	forecast, err := svc.Forecast(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forecast.WeightedPipeline != 0 {
		t.Errorf("weighted_pipeline = %v, want 0", forecast.WeightedPipeline)
	}
	if forecast.WinRate != 0 {
		t.Errorf("win_rate = %v, want 0", forecast.WinRate)
	}
}

func TestPipelineService_Forecast_GroupsByStageInSequenceOrder(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{
				{Base: model.Base{ID: 1}, Name: "New", Sequence: 10, Probability: 10},
				{Base: model.Base{ID: 2}, Name: "Qualified", Sequence: 20, Probability: 35},
				{Base: model.Base{ID: 3}, Name: "Proposal Sent", Sequence: 30, Probability: 60},
			}}, nil
		},
	}
	leads := ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*Prospect, error) {
			return []*Prospect{
				{StageID: &[]uint64{3}[0], ExpectedRevenue: 5000, Probability: 60},
				{StageID: &[]uint64{1}[0], ExpectedRevenue: 1000, Probability: 10},
			}, nil
		},
	}
	svc := NewPipelineService(leads, stages)

	forecast, err := svc.Forecast(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forecast.Stages) != 3 {
		t.Fatalf("len(stages) = %d, want 3", len(forecast.Stages))
	}
	if forecast.Stages[0].StageName != "New" || forecast.Stages[0].OpportunityCount != 1 || forecast.Stages[0].WeightedRevenue != 100 {
		t.Errorf("stage[0] = %+v, want New/count 1/weighted 100", forecast.Stages[0])
	}
	if forecast.Stages[1].OpportunityCount != 0 || forecast.Stages[1].WeightedRevenue != 0 {
		t.Errorf("stage[1] = %+v, want empty stage", forecast.Stages[1])
	}
	if forecast.Stages[2].StageName != "Proposal Sent" || forecast.Stages[2].OpportunityCount != 1 || forecast.Stages[2].WeightedRevenue != 3000 {
		t.Errorf("stage[2] = %+v, want Proposal Sent/count 1/weighted 3000", forecast.Stages[2])
	}
}

func TestPipelineService_Forecast_PropagatesStageListError(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewPipelineService(ProspectDAOMock{}, stages)

	_, err := svc.Forecast(ctx, nil)

	helper.AssertError(t, err, true, nil)
}

func TestPipelineService_Forecast_PropagatesListOpenError(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{}}, nil
		},
	}
	leads := ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*Prospect, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewPipelineService(leads, stages)

	_, err := svc.Forecast(ctx, nil)

	helper.AssertError(t, err, true, nil)
}

func TestPipelineService_Forecast_PropagatesCountWonError(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{}}, nil
		},
	}
	leads := ProspectDAOMock{
		CountWonFunc: func(_ context.Context, _ *uint64) (int64, error) {
			return 0, errors.New("db down")
		},
	}
	svc := NewPipelineService(leads, stages)

	_, err := svc.Forecast(ctx, nil)

	helper.AssertError(t, err, true, nil)
}

func TestPipelineService_Forecast_PropagatesCountLostError(t *testing.T) {
	ctx := context.Background()
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{}}, nil
		},
	}
	leads := ProspectDAOMock{
		CountLostFunc: func(_ context.Context, _ *uint64) (int64, error) {
			return 0, errors.New("db down")
		},
	}
	svc := NewPipelineService(leads, stages)

	_, err := svc.Forecast(ctx, nil)

	helper.AssertError(t, err, true, nil)
}

func TestPipelineService_Forecast_ScopedToOrganization(t *testing.T) {
	ctx := context.Background()
	orgID := helper.Ptr(uint64(10))
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.PipelineStage], error) {
			if len(q.Filters) != 1 || q.Filters[0].Value != uint64(10) {
				t.Errorf("filters = %+v, want organization_id filter", q.Filters)
			}
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{
				{Base: model.Base{ID: 1}, Name: "New", Probability: 10},
			}}, nil
		},
	}
	leads := ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*Prospect, error) {
			return []*Prospect{
				{StageID: &[]uint64{1}[0], ExpectedRevenue: 1000, Probability: 10},
			}, nil
		},
	}
	svc := NewPipelineService(leads, stages)

	forecast, err := svc.Forecast(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forecast.WeightedPipeline != 100 {
		t.Errorf("weighted_pipeline = %v, want 100", forecast.WeightedPipeline)
	}
}
