package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type MaintenanceService struct {
	plans      MaintenancePlanDAO
	equipments EquipmentDAO
	orders     ServiceOrderDAO
	tx         db.Transactioner
}

func NewMaintenanceService(
	plans MaintenancePlanDAO,
	equipments EquipmentDAO,
	orders ServiceOrderDAO,
	tx db.Transactioner,
) MaintenanceService {
	return MaintenanceService{plans: plans, equipments: equipments, orders: orders, tx: tx}
}

type CreatePlanRequest struct {
	EquipmentID  uint64
	Name         string
	IntervalDays int
	NextDue      *time.Time
}

func (s MaintenanceService) ListPlans(ctx context.Context, q *query.Query) (*query.Page[MaintenancePlan], error) {
	return s.plans.List(ctx, q)
}

func (s MaintenanceService) FindPlan(ctx context.Context, id uint64) (*MaintenancePlan, error) {
	return s.plans.Find(ctx, id)
}

func (s MaintenanceService) CreatePlan(ctx context.Context, request CreatePlanRequest) (*MaintenancePlan, error) {
	if request.Name == "" {
		return nil, ErrPlanNameRequired
	}
	if request.IntervalDays <= 0 {
		return nil, ErrPlanIntervalInvalid
	}
	if request.NextDue == nil {
		return nil, ErrPlanNextDueRequired
	}
	equipment, err := s.equipments.Find(ctx, request.EquipmentID)
	if err != nil {
		return nil, err
	}
	if equipment == nil {
		return nil, ErrEquipmentNotFound
	}
	return s.plans.Create(ctx, &MaintenancePlan{
		EquipmentID:  &request.EquipmentID,
		Name:         request.Name,
		IntervalDays: request.IntervalDays,
		NextDue:      request.NextDue,
		Active:       true,
	})
}

func (s MaintenanceService) GenerateDueOrdersNow(ctx context.Context, organizationID uint64) ([]*ServiceOrder, error) {
	return s.GenerateDueOrders(ctx, organizationID, time.Now().UTC())
}

func (s MaintenanceService) GenerateDueOrders(ctx context.Context, organizationID uint64, asOf time.Time) ([]*ServiceOrder, error) {
	due, err := s.plans.ListDue(ctx, organizationID, asOf)
	if err != nil {
		return nil, err
	}
	var generated []*ServiceOrder
	for _, plan := range due {
		equipment, err := s.equipments.Find(ctx, helper.Deref(plan.EquipmentID, 0))
		if err != nil {
			return nil, err
		}
		if equipment == nil || equipment.OrganizationID == nil || plan.NextDue == nil {
			continue
		}
		dueDate := plan.NextDue
		name := plan.Name
		if name == "" {
			name = "Preventive maintenance"
		}
		name = fmt.Sprintf("%s (%s)", name, dueDate.Format("2006-01-02"))
		var order *ServiceOrder
		err = s.tx.Run(ctx, func(tx *gorm.DB) error {
			createdOrder, err := s.orders.CreateTx(ctx, tx, &ServiceOrder{
				OrganizationID: equipment.OrganizationID,
				Name:           name,
				EquipmentID:    plan.EquipmentID,
				Type:           OrderTypeMaintenance,
				State:          OrderStateScheduled,
				ScheduledDate:  dueDate,
			})
			if err != nil {
				return err
			}
			next := advanceNextDue(*dueDate, plan.IntervalDays, asOf)
			plan.NextDue = &next
			if _, err := s.plans.UpdateTx(ctx, tx, plan); err != nil {
				return err
			}
			order = createdOrder
			return nil
		})
		if err != nil {
			return nil, err
		}
		generated = append(generated, order)
	}
	return generated, nil
}

func advanceNextDue(due time.Time, intervalDays int, asOf time.Time) time.Time {
	for !due.After(asOf) {
		due = due.AddDate(0, 0, intervalDays)
	}
	return due
}
