package project

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type ProjectDAOMock struct {
	dao.CRUDMock[Project]
}

type ProjectTaskDAOMock struct {
	dao.CRUDMock[ProjectTask]
	ListByProjectFunc func(ctx context.Context, projectID uint64) ([]*ProjectTask, error)
}

func (m ProjectTaskDAOMock) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectTask, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []*ProjectTask{}, nil
}

type ProjectMilestoneDAOMock struct {
	dao.CRUDMock[ProjectMilestone]
	ListByProjectFunc func(ctx context.Context, projectID uint64) ([]*ProjectMilestone, error)
}

func (m ProjectMilestoneDAOMock) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectMilestone, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []*ProjectMilestone{}, nil
}

type ProjectInvoiceLineDAOMock struct {
	dao.CRUDMock[ProjectInvoiceLine]
	ListByProjectFunc   func(ctx context.Context, projectID uint64) ([]*ProjectInvoiceLine, error)
	ListByTimesheetFunc func(ctx context.Context, timesheetID uint64) ([]*ProjectInvoiceLine, error)
}

func (m ProjectInvoiceLineDAOMock) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectInvoiceLine, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []*ProjectInvoiceLine{}, nil
}

func (m ProjectInvoiceLineDAOMock) ListByTimesheet(ctx context.Context, timesheetID uint64) ([]*ProjectInvoiceLine, error) {
	if m.ListByTimesheetFunc != nil {
		return m.ListByTimesheetFunc(ctx, timesheetID)
	}
	return []*ProjectInvoiceLine{}, nil
}
