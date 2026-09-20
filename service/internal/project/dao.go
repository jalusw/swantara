package project

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ProjectDAO interface {
	dao.CRUD[Project]
}

type projectDAO struct {
	dao.Base[Project]
}

func NewProjectDAO(db *gorm.DB) ProjectDAO {
	return projectDAO{Base: dao.NewBase[Project](db)}
}

type ProjectTaskDAO interface {
	dao.CRUD[ProjectTask]
	ListByProject(ctx context.Context, projectID uint64) ([]*ProjectTask, error)
}

type projectTaskDAO struct {
	dao.Base[ProjectTask]
}

func NewProjectTaskDAO(db *gorm.DB) ProjectTaskDAO {
	return projectTaskDAO{Base: dao.NewBase[ProjectTask](db)}
}

func (d projectTaskDAO) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectTask, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "project_id", Operator: query.Equal, Value: projectID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ProjectMilestoneDAO interface {
	dao.CRUD[ProjectMilestone]
	ListByProject(ctx context.Context, projectID uint64) ([]*ProjectMilestone, error)
}

type projectMilestoneDAO struct {
	dao.Base[ProjectMilestone]
}

func NewProjectMilestoneDAO(db *gorm.DB) ProjectMilestoneDAO {
	return projectMilestoneDAO{Base: dao.NewBase[ProjectMilestone](db)}
}

func (d projectMilestoneDAO) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectMilestone, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "project_id", Operator: query.Equal, Value: projectID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ProjectInvoiceLineDAO interface {
	dao.CRUD[ProjectInvoiceLine]
	ListByProject(ctx context.Context, projectID uint64) ([]*ProjectInvoiceLine, error)
	ListByTimesheet(ctx context.Context, timesheetID uint64) ([]*ProjectInvoiceLine, error)
}

type projectInvoiceLineDAO struct {
	dao.Base[ProjectInvoiceLine]
}

func NewProjectInvoiceLineDAO(db *gorm.DB) ProjectInvoiceLineDAO {
	return projectInvoiceLineDAO{Base: dao.NewBase[ProjectInvoiceLine](db)}
}

func (d projectInvoiceLineDAO) ListByProject(ctx context.Context, projectID uint64) ([]*ProjectInvoiceLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "project_id", Operator: query.Equal, Value: projectID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d projectInvoiceLineDAO) ListByTimesheet(ctx context.Context, timesheetID uint64) ([]*ProjectInvoiceLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "timesheet_id", Operator: query.Equal, Value: timesheetID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
