package service

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type MaintenancePlanDAO interface {
	dao.CRUD[MaintenancePlan]
	ListDue(ctx context.Context, organizationID uint64, asOf time.Time) ([]*MaintenancePlan, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *MaintenancePlan) (*MaintenancePlan, error)
}

type maintenancePlanDAO struct {
	dao.Base[MaintenancePlan]
	db *gorm.DB
}

func NewMaintenancePlanDAO(db *gorm.DB) MaintenancePlanDAO {
	return maintenancePlanDAO{Base: dao.NewBase[MaintenancePlan](db), db: db}
}

func (d maintenancePlanDAO) ListDue(ctx context.Context, organizationID uint64, asOf time.Time) ([]*MaintenancePlan, error) {
	var plans []MaintenancePlan
	err := d.db.WithContext(ctx).
		Joins("JOIN equipments ON equipments.id = maintenance_plans.equipment_id").
		Where("equipments.organization_id = ? AND maintenance_plans.active = ? AND maintenance_plans.next_due IS NOT NULL AND maintenance_plans.next_due <= ? AND maintenance_plans.deleted_at IS NULL", organizationID, true, asOf).
		Order("maintenance_plans.next_due").
		Find(&plans).Error
	if err != nil {
		return nil, err
	}
	result := make([]*MaintenancePlan, 0, len(plans))
	for i := range plans {
		result = append(result, &plans[i])
	}
	return result, nil
}

func (d maintenancePlanDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *MaintenancePlan) (*MaintenancePlan, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}
