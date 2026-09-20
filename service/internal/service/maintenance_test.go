package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestMaintenanceService_GenerateDueOrders_CreatesOneOrderPerDuePlan(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	dueDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	plan := &MaintenancePlan{Base: model.Base{ID: 1}, EquipmentID: helper.Ptr(uint64(5)), Name: "Quarterly service", IntervalDays: 30, NextDue: &dueDate, Active: true}
	equipment := &Equipment{Base: model.Base{ID: 5}, OrganizationID: &organizationID, Name: "Chiller"}
	createdOrders := []*ServiceOrder{}
	svc := NewMaintenanceService(
		MaintenancePlanDAOMock{
			ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
				return []*MaintenancePlan{plan}, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, updated *MaintenancePlan) (*MaintenancePlan, error) {
				plan = updated
				return updated, nil
			},
		},
		EquipmentDAOMock{FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) { return equipment, nil }},
		ServiceOrderDAOMock{CreateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			order.ID = uint64(len(createdOrders) + 1)
			createdOrders = append(createdOrders, order)
			return order, nil
		}},
		txMock{},
	)

	generated, err := svc.GenerateDueOrders(ctx, organizationID, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(generated) != 1 {
		t.Fatalf("generated = %d, want 1", len(generated))
	}
	order := createdOrders[0]
	if order.Type != OrderTypeMaintenance || order.State != OrderStateScheduled {
		t.Errorf("order = %+v, want maintenance / scheduled", order)
	}
	if order.EquipmentID == nil || *order.EquipmentID != 5 {
		t.Errorf("order equipment = %v, want 5", order.EquipmentID)
	}
	if order.OrganizationID == nil || *order.OrganizationID != 10 {
		t.Errorf("order organization = %v, want 10", order.OrganizationID)
	}
	if order.ScheduledDate == nil || !order.ScheduledDate.Equal(dueDate) {
		t.Errorf("order scheduled date = %v, want %v", order.ScheduledDate, dueDate)
	}
	if plan.NextDue == nil || !plan.NextDue.After(asOf) {
		t.Errorf("plan next due = %v, want advanced past %v", plan.NextDue, asOf)
	}
}

func TestMaintenanceService_GenerateDueOrders_IsIdempotentOnReRun(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	dueDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	plan := &MaintenancePlan{Base: model.Base{ID: 1}, EquipmentID: helper.Ptr(uint64(5)), Name: "Quarterly service", IntervalDays: 30, NextDue: &dueDate, Active: true}
	equipment := &Equipment{Base: model.Base{ID: 5}, OrganizationID: &organizationID, Name: "Chiller"}
	createdOrders := []*ServiceOrder{}
	svc := NewMaintenanceService(
		MaintenancePlanDAOMock{
			ListDueFunc: func(_ context.Context, _ uint64, asOf time.Time) ([]*MaintenancePlan, error) {
				if plan.NextDue != nil && plan.NextDue.After(asOf) {
					return []*MaintenancePlan{}, nil
				}
				return []*MaintenancePlan{plan}, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, updated *MaintenancePlan) (*MaintenancePlan, error) {
				plan = updated
				return updated, nil
			},
		},
		EquipmentDAOMock{FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) { return equipment, nil }},
		ServiceOrderDAOMock{CreateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			order.ID = uint64(len(createdOrders) + 1)
			createdOrders = append(createdOrders, order)
			return order, nil
		}},
		txMock{},
	)

	if _, err := svc.GenerateDueOrders(ctx, organizationID, asOf); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if _, err := svc.GenerateDueOrders(ctx, organizationID, asOf); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(createdOrders) != 1 {
		t.Errorf("orders = %d, want 1 (idempotent re-run)", len(createdOrders))
	}
	if plan.NextDue == nil || !plan.NextDue.After(asOf) {
		t.Errorf("plan next due = %v, want advanced past %v after re-run", plan.NextDue, asOf)
	}
}

func TestMaintenanceService_CreatePlan_Validates(t *testing.T) {
	svc := NewMaintenanceService(MaintenancePlanDAOMock{}, EquipmentDAOMock{}, ServiceOrderDAOMock{}, txMock{})
	ctx := context.Background()
	next := time.Now()

	if _, err := svc.CreatePlan(ctx, CreatePlanRequest{Name: "", IntervalDays: 30, NextDue: &next}); !errors.Is(err, ErrPlanNameRequired) {
		t.Errorf("empty name err = %v, want ErrPlanNameRequired", err)
	}
	if _, err := svc.CreatePlan(ctx, CreatePlanRequest{Name: "Plan", IntervalDays: 0, NextDue: &next}); !errors.Is(err, ErrPlanIntervalInvalid) {
		t.Errorf("invalid interval err = %v, want ErrPlanIntervalInvalid", err)
	}
	if _, err := svc.CreatePlan(ctx, CreatePlanRequest{Name: "Plan", IntervalDays: 30, NextDue: nil}); !errors.Is(err, ErrPlanNextDueRequired) {
		t.Errorf("nil next due err = %v, want ErrPlanNextDueRequired", err)
	}
	if _, err := svc.CreatePlan(ctx, CreatePlanRequest{Name: "Plan", IntervalDays: 30, NextDue: &next}); !errors.Is(err, ErrEquipmentNotFound) {
		t.Errorf("missing equipment err = %v, want ErrEquipmentNotFound", err)
	}
}
