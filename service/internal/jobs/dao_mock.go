package jobs

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type DAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[JobRun], error)
	SearchFunc     func(ctx context.Context, field string, value any) (*JobRun, error)
	FindFunc       func(ctx context.Context, id uint64) (*JobRun, error)
	CreateFunc     func(ctx context.Context, run *JobRun) (*JobRun, error)
	UpdateFunc     func(ctx context.Context, run *JobRun) (*JobRun, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
}

func (m DAOMock) List(ctx context.Context, q *query.Query) (*query.Page[JobRun], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[JobRun]{Items: []*JobRun{}, Count: 0}, nil
}

func (m DAOMock) Search(ctx context.Context, field string, value any) (*JobRun, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return nil, nil
}

func (m DAOMock) Find(ctx context.Context, id uint64) (*JobRun, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m DAOMock) Create(ctx context.Context, run *JobRun) (*JobRun, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, run)
	}
	return run, nil
}

func (m DAOMock) Update(ctx context.Context, run *JobRun) (*JobRun, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, run)
	}
	return run, nil
}

func (m DAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m DAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}
