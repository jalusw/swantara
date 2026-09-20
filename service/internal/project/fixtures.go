package project

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ProjectFixture(opts ...func(*Project) *Project) *Project {
	now := time.Now()
	p := &Project{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		Name:           gofakeit.AppName(),
		ContactID:      uint64(gofakeit.Number(1, 10000)),
		ManagerID:      nil,
		DimensionID:    nil,
		SaleOrderID:    nil,
		BillingType:    "manual",
		BillableRate:   gofakeit.Float64Range(1, 10000),
		DateStart:      helper.Ptr(now),
		DateEnd:        nil,
		State:          "draft",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func ProjectTaskFixture(opts ...func(*ProjectTask) *ProjectTask) *ProjectTask {
	now := time.Now()
	t := &ProjectTask{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ProjectID:      uint64(gofakeit.Number(1, 10000)),
		Name:           gofakeit.AppName(),
		AssigneeID:     nil,
		Stage:          "in_progress",
		PlannedHours:   gofakeit.Float64Range(1, 100),
		EffectiveHours: gofakeit.Float64Range(0, 100),
		ParentTaskID:   nil,
		Deadline:       nil,
		Priority:       helper.Ptr(1),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func ProjectMilestoneFixture(opts ...func(*ProjectMilestone) *ProjectMilestone) *ProjectMilestone {
	now := time.Now()
	m := &ProjectMilestone{
		Base:       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ProjectID:  uint64(gofakeit.Number(1, 10000)),
		Name:       gofakeit.AppName(),
		Deadline:   nil,
		Reached:    false,
		SaleLineID: nil,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func ProjectInvoiceLineFixture(opts ...func(*ProjectInvoiceLine) *ProjectInvoiceLine) *ProjectInvoiceLine {
	now := time.Now()
	l := &ProjectInvoiceLine{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ProjectID:     uint64(gofakeit.Number(1, 10000)),
		InvoiceID:     uint64(gofakeit.Number(1, 10000)),
		InvoiceLineID: nil,
		TimesheetID:   uint64(gofakeit.Number(1, 10000)),
		Qty:           gofakeit.Float64Range(1, 100),
		UnitPrice:     gofakeit.Float64Range(1, 10000),
		Amount:        gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}
