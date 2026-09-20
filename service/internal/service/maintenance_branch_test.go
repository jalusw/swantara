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

func TestMaintenanceService_CreatePlan_PropagatesEquipmentFindError(t *testing.T) {
	ctx := context.Background()
	next := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	equipments := EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestMaintenanceService(MaintenancePlanDAOMock{}, equipments, ServiceOrderDAOMock{}, ServiceTransactionerMock{})

	_, err := svc.CreatePlan(ctx, CreatePlanRequest{EquipmentID: 5, Name: "Monthly", IntervalDays: 30, NextDue: &next})
	if err == nil {
		t.Error("expected equipment find error to propagate")
	}
}

func TestMaintenanceService_GenerateDueOrders_PropagatesPlanUpdateError(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	plan := &MaintenancePlan{Base: model.Base{ID: 1}, EquipmentID: helper.Ptr(uint64(5)), Name: "Quarterly", IntervalDays: 30, NextDue: &due}

	plans := MaintenancePlanDAOMock{
		ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*MaintenancePlan, error) {
			return []*MaintenancePlan{plan}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *MaintenancePlan) (*MaintenancePlan, error) {
			return nil, errors.New("update failed")
		},
	}
	equipments := EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*Equipment, error) {
			return &Equipment{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	svc := NewTestMaintenanceService(plans, equipments, ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			order.ID = 1
			return order, nil
		},
	}, ServiceTransactionerMock{})

	_, err := svc.GenerateDueOrders(ctx, 10, asOf)
	if err == nil {
		t.Error("expected plan update error to propagate")
	}
}
