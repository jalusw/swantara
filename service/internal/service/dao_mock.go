package service

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type EquipmentDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[Equipment], error)
	FindFunc       func(ctx context.Context, id uint64) (*Equipment, error)
	CreateFunc     func(ctx context.Context, entity *Equipment) (*Equipment, error)
	UpdateFunc     func(ctx context.Context, entity *Equipment) (*Equipment, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
}

func (m EquipmentDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[Equipment], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[Equipment]{}, nil
}

func (m EquipmentDAOMock) Find(ctx context.Context, id uint64) (*Equipment, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m EquipmentDAOMock) Create(ctx context.Context, entity *Equipment) (*Equipment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m EquipmentDAOMock) Update(ctx context.Context, entity *Equipment) (*Equipment, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m EquipmentDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m EquipmentDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m EquipmentDAOMock) Search(ctx context.Context, field string, value any) (*Equipment, error) {
	return nil, nil
}

type ServiceContractDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[ServiceContract], error)
	FindFunc       func(ctx context.Context, id uint64) (*ServiceContract, error)
	CreateFunc     func(ctx context.Context, entity *ServiceContract) (*ServiceContract, error)
	UpdateFunc     func(ctx context.Context, entity *ServiceContract) (*ServiceContract, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	UpdateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *ServiceContract) (*ServiceContract, error)
}

func (m ServiceContractDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[ServiceContract], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[ServiceContract]{}, nil
}

func (m ServiceContractDAOMock) Find(ctx context.Context, id uint64) (*ServiceContract, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m ServiceContractDAOMock) Create(ctx context.Context, entity *ServiceContract) (*ServiceContract, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceContractDAOMock) Update(ctx context.Context, entity *ServiceContract) (*ServiceContract, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceContractDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceContractDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceContractDAOMock) Search(ctx context.Context, field string, value any) (*ServiceContract, error) {
	return nil, nil
}

func (m ServiceContractDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceContract) (*ServiceContract, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

type ServiceOrderDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[ServiceOrder], error)
	FindFunc       func(ctx context.Context, id uint64) (*ServiceOrder, error)
	CreateFunc     func(ctx context.Context, entity *ServiceOrder) (*ServiceOrder, error)
	UpdateFunc     func(ctx context.Context, entity *ServiceOrder) (*ServiceOrder, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	UpdateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error)
}

func (m ServiceOrderDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[ServiceOrder], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[ServiceOrder]{}, nil
}

func (m ServiceOrderDAOMock) Find(ctx context.Context, id uint64) (*ServiceOrder, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m ServiceOrderDAOMock) Create(ctx context.Context, entity *ServiceOrder) (*ServiceOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceOrderDAOMock) Update(ctx context.Context, entity *ServiceOrder) (*ServiceOrder, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceOrderDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceOrderDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceOrderDAOMock) Search(ctx context.Context, field string, value any) (*ServiceOrder, error) {
	return nil, nil
}

func (m ServiceOrderDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

type ServiceOrderLineDAOMock struct {
	ListFunc        func(ctx context.Context, q *query.Query) (*query.Page[ServiceOrderLine], error)
	FindFunc        func(ctx context.Context, id uint64) (*ServiceOrderLine, error)
	CreateFunc      func(ctx context.Context, entity *ServiceOrderLine) (*ServiceOrderLine, error)
	UpdateFunc      func(ctx context.Context, entity *ServiceOrderLine) (*ServiceOrderLine, error)
	DeleteFunc      func(ctx context.Context, id uint64) error
	HardDeleteFunc  func(ctx context.Context, id uint64) error
	CreateTxFunc    func(ctx context.Context, tx *gorm.DB, entity *ServiceOrderLine) (*ServiceOrderLine, error)
	ListByOrderFunc func(ctx context.Context, orderID uint64) ([]*ServiceOrderLine, error)
}

func (m ServiceOrderLineDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[ServiceOrderLine], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[ServiceOrderLine]{}, nil
}

func (m ServiceOrderLineDAOMock) Find(ctx context.Context, id uint64) (*ServiceOrderLine, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m ServiceOrderLineDAOMock) Create(ctx context.Context, entity *ServiceOrderLine) (*ServiceOrderLine, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceOrderLineDAOMock) Update(ctx context.Context, entity *ServiceOrderLine) (*ServiceOrderLine, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m ServiceOrderLineDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceOrderLineDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m ServiceOrderLineDAOMock) Search(ctx context.Context, field string, value any) (*ServiceOrderLine, error) {
	return nil, nil
}

func (m ServiceOrderLineDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrderLine) (*ServiceOrderLine, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m ServiceOrderLineDAOMock) ListByOrder(ctx context.Context, orderID uint64) ([]*ServiceOrderLine, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return []*ServiceOrderLine{}, nil
}

type MaintenancePlanDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[MaintenancePlan], error)
	FindFunc       func(ctx context.Context, id uint64) (*MaintenancePlan, error)
	CreateFunc     func(ctx context.Context, entity *MaintenancePlan) (*MaintenancePlan, error)
	UpdateFunc     func(ctx context.Context, entity *MaintenancePlan) (*MaintenancePlan, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	ListDueFunc    func(ctx context.Context, organizationID uint64, asOf time.Time) ([]*MaintenancePlan, error)
	UpdateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *MaintenancePlan) (*MaintenancePlan, error)
}

func (m MaintenancePlanDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[MaintenancePlan], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[MaintenancePlan]{}, nil
}

func (m MaintenancePlanDAOMock) Find(ctx context.Context, id uint64) (*MaintenancePlan, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m MaintenancePlanDAOMock) Create(ctx context.Context, entity *MaintenancePlan) (*MaintenancePlan, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m MaintenancePlanDAOMock) Update(ctx context.Context, entity *MaintenancePlan) (*MaintenancePlan, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m MaintenancePlanDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m MaintenancePlanDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m MaintenancePlanDAOMock) Search(ctx context.Context, field string, value any) (*MaintenancePlan, error) {
	return nil, nil
}

func (m MaintenancePlanDAOMock) ListDue(ctx context.Context, organizationID uint64, asOf time.Time) ([]*MaintenancePlan, error) {
	if m.ListDueFunc != nil {
		return m.ListDueFunc(ctx, organizationID, asOf)
	}
	return []*MaintenancePlan{}, nil
}

func (m MaintenancePlanDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *MaintenancePlan) (*MaintenancePlan, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}
