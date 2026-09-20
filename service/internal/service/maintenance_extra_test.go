package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestMaintenanceService_CreatePlan_Success(t *testing.T) {
	ctx := context.Background()
	next := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	var created *MaintenancePlan
	plans := MaintenancePlanDAOMock{
		CreateFunc: func(_ context.Context, plan *MaintenancePlan) (*MaintenancePlan, error) {
			created = plan
			return plan, nil
		},
	}
	equipments := EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) {
			return &Equipment{Base: model.Base{ID: 5}}, nil
		},
	}
	svc := NewMaintenanceService(plans, equipments, ServiceOrderDAOMock{}, txMock{})

	plan, err := svc.CreatePlan(ctx, CreatePlanRequest{EquipmentID: 5, Name: "Monthly", IntervalDays: 30, NextDue: &next})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.Active || plan.IntervalDays != 30 || created == nil {
		t.Errorf("plan = %+v, created = %v", plan, created)
	}
}

func TestMaintenanceService_GenerateDueOrders_PropagatesErrors(t *testing.T) {
	ctx := context.Background()
	asOf := time.Now()

	listErr := NewMaintenanceService(
		MaintenancePlanDAOMock{ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
			return nil, errors.New("list failed")
		}},
		EquipmentDAOMock{},
		ServiceOrderDAOMock{},
		txMock{},
	)
	if _, err := listErr.GenerateDueOrders(ctx, 10, asOf); err == nil {
		t.Error("expected ListDue error propagation")
	}

	due := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	plan := &MaintenancePlan{Base: model.Base{ID: 1}, EquipmentID: helper.Ptr(uint64(5)), IntervalDays: 30, NextDue: &due}
	equipErr := NewMaintenanceService(
		MaintenancePlanDAOMock{ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
			return []*MaintenancePlan{plan}, nil
		}},
		EquipmentDAOMock{FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) {
			return nil, errors.New("equipment failed")
		}},
		ServiceOrderDAOMock{},
		txMock{},
	)
	if _, err := equipErr.GenerateDueOrders(ctx, 10, asOf); err == nil {
		t.Error("expected equipment error propagation")
	}

	orderErr := NewMaintenanceService(
		MaintenancePlanDAOMock{ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
			return []*MaintenancePlan{plan}, nil
		}},
		EquipmentDAOMock{FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) {
			return &Equipment{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(10))}, nil
		}},
		ServiceOrderDAOMock{CreateFunc: func(_ context.Context, _ *ServiceOrder) (*ServiceOrder, error) {
			return nil, errors.New("create failed")
		}},
		txMock{},
	)
	if _, err := orderErr.GenerateDueOrders(ctx, 10, asOf); err == nil {
		t.Error("expected order create error propagation")
	}
}

func TestMaintenanceService_GenerateDueOrders_SkipsUnroutablePlans(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	dueDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	noEquipment := &MaintenancePlan{Base: model.Base{ID: 1}, EquipmentID: helper.Ptr(uint64(5)), NextDue: &dueDate}
	noOrg := &MaintenancePlan{Base: model.Base{ID: 2}, EquipmentID: helper.Ptr(uint64(6)), NextDue: &dueDate}
	noDue := &MaintenancePlan{Base: model.Base{ID: 3}, EquipmentID: helper.Ptr(uint64(7))}

	svc := NewMaintenanceService(
		MaintenancePlanDAOMock{ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
			return []*MaintenancePlan{noEquipment, noOrg, noDue}, nil
		}},
		EquipmentDAOMock{FindFunc: func(_ context.Context, equipmentID uint64) (*Equipment, error) {
			switch equipmentID {
			case 5:
				return nil, nil
			case 6:
				return &Equipment{Base: model.Base{ID: 6}}, nil
			default:
				return &Equipment{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
			}
		}},
		ServiceOrderDAOMock{CreateFunc: func(_ context.Context, _ *ServiceOrder) (*ServiceOrder, error) {
			return nil, nil
		}},
		txMock{},
	)

	generated, err := svc.GenerateDueOrders(ctx, 10, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(generated) != 0 {
		t.Errorf("generated = %d, want 0 (all plans unroutable)", len(generated))
	}
}

func TestAdvanceNextDue(t *testing.T) {
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	advanced := advanceNextDue(due, 30, asOf)
	if !advanced.After(asOf) {
		t.Errorf("advanced = %v, want after %v", advanced, asOf)
	}
	if !advanced.Equal(time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("advanced = %v, want 2026-08-14 (6/15 + 30 + 30)", advanced)
	}

	alreadyFuture := advanceNextDue(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), 30, asOf)
	if !alreadyFuture.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("alreadyFuture = %v, want unchanged", alreadyFuture)
	}
}
