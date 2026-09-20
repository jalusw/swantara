package service

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func EquipmentFixture(opts ...func(*Equipment) *Equipment) *Equipment {
	e := &Equipment{
		Base: model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		Name: gofakeit.Name(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func ServiceOrderFixture(opts ...func(*ServiceOrder) *ServiceOrder) *ServiceOrder {
	o := &ServiceOrder{
		Base:  model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		Name:  gofakeit.Name(),
		Type:  OrderTypeRepair,
		State: OrderStateNew,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
