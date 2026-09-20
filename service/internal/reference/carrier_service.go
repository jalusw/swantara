package reference

import (
	"context"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type VariantExistenceChecker interface {
	VariantExists(ctx context.Context, id uint64) (bool, error)
}

type CarrierService struct {
	carriers dao.Base[Carrier]
	variants VariantExistenceChecker
}

func NewCarrierService(carriers dao.Base[Carrier], variants VariantExistenceChecker) CarrierService {
	return CarrierService{carriers: carriers, variants: variants}
}

func (s CarrierService) List(ctx context.Context, q *query.Query) (*query.Page[Carrier], error) {
	return s.carriers.List(ctx, q)
}

func (s CarrierService) Find(ctx context.Context, id uint64) (*Carrier, error) {
	return s.carriers.Find(ctx, id)
}

func (s CarrierService) Delete(ctx context.Context, id uint64) error {
	return s.carriers.Delete(ctx, id)
}

func (s CarrierService) Create(ctx context.Context, carrier *Carrier) (*Carrier, error) {
	if err := s.validate(ctx, carrier); err != nil {
		return nil, err
	}
	return s.carriers.Create(ctx, carrier)
}

func (s CarrierService) Update(ctx context.Context, carrier *Carrier) (*Carrier, error) {
	if err := s.validate(ctx, carrier); err != nil {
		return nil, err
	}
	return s.carriers.Update(ctx, carrier)
}

func (s CarrierService) validate(ctx context.Context, carrier *Carrier) error {
	if strings.TrimSpace(carrier.Name) == "" {
		return ErrCarrierName
	}
	if carrier.DeliveryItemID != nil {
		if s.variants == nil {
			return nil
		}
		exists, err := s.variants.VariantExists(ctx, *carrier.DeliveryItemID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrCarrierDeliveryProduct
		}
	}
	return nil
}
