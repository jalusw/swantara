package service

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type EquipmentDAO interface {
	dao.CRUD[Equipment]
}

type equipmentDAO struct {
	dao.Base[Equipment]
}

func NewEquipmentDAO(db *gorm.DB) EquipmentDAO {
	return equipmentDAO{Base: dao.NewBase[Equipment](db)}
}

type ServiceContractDAO interface {
	dao.CRUD[ServiceContract]
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceContract) (*ServiceContract, error)
}

type serviceContractDAO struct {
	dao.Base[ServiceContract]
	db *gorm.DB
}

func NewServiceContractDAO(db *gorm.DB) ServiceContractDAO {
	return serviceContractDAO{Base: dao.NewBase[ServiceContract](db), db: db}
}

func (d serviceContractDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceContract) (*ServiceContract, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

type ServiceOrderDAO interface {
	dao.CRUD[ServiceOrder]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error)
}

type serviceOrderDAO struct {
	dao.Base[ServiceOrder]
	db *gorm.DB
}

func NewServiceOrderDAO(db *gorm.DB) ServiceOrderDAO {
	return serviceOrderDAO{Base: dao.NewBase[ServiceOrder](db), db: db}
}

func (d serviceOrderDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d serviceOrderDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrder) (*ServiceOrder, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

type ServiceOrderLineDAO interface {
	dao.CRUD[ServiceOrderLine]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrderLine) (*ServiceOrderLine, error)
	ListByOrder(ctx context.Context, orderID uint64) ([]*ServiceOrderLine, error)
}

type serviceOrderLineDAO struct {
	dao.Base[ServiceOrderLine]
	db *gorm.DB
}

func NewServiceOrderLineDAO(db *gorm.DB) ServiceOrderLineDAO {
	return serviceOrderLineDAO{Base: dao.NewBase[ServiceOrderLine](db), db: db}
}

func (d serviceOrderLineDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *ServiceOrderLine) (*ServiceOrderLine, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d serviceOrderLineDAO) ListByOrder(ctx context.Context, orderID uint64) ([]*ServiceOrderLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "service_order_id", Operator: query.Equal, Value: orderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
