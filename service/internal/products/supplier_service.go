package products

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type SupplierProductService struct {
	variants         ItemVariantDAO
	templates        ItemDAO
	supplierProducts SupplierProductDAO
	contactDAO       contacts.ContactDAO
	supplierDAO      contacts.SupplierProfileDAO
}

func NewSupplierProductService(
	variants ItemVariantDAO,
	templates ItemDAO,
	supplierProducts SupplierProductDAO,
	contactDAO contacts.ContactDAO,
	supplierDAO contacts.SupplierProfileDAO,
) SupplierProductService {
	return SupplierProductService{
		variants:         variants,
		templates:        templates,
		supplierProducts: supplierProducts,
		contactDAO:       contactDAO,
		supplierDAO:      supplierDAO,
	}
}

func (s SupplierProductService) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[SupplierProduct], error) {
	return s.supplierProducts.ListInOrg(ctx, q, organizationID)
}

func (s SupplierProductService) FindInOrg(ctx context.Context, id, organizationID uint64) (*SupplierProduct, error) {
	return s.supplierProducts.FindInOrg(ctx, id, organizationID)
}

func (s SupplierProductService) Delete(ctx context.Context, id uint64) error {
	return s.supplierProducts.Delete(ctx, id)
}

func (s SupplierProductService) Create(ctx context.Context, organizationID uint64, offer *SupplierProduct) (*SupplierProduct, error) {
	if err := s.validateOffer(ctx, organizationID, offer); err != nil {
		return nil, err
	}
	if offer.MinQty == 0 {
		offer.MinQty = 1
	}
	if offer.Priority == 0 {
		offer.Priority = DefaultSupplierPriority
	}
	return s.supplierProducts.Create(ctx, offer)
}

func (s SupplierProductService) Update(ctx context.Context, organizationID uint64, offer *SupplierProduct) (*SupplierProduct, error) {
	if err := s.validateOffer(ctx, organizationID, offer); err != nil {
		return nil, err
	}
	return s.supplierProducts.Update(ctx, offer)
}

func (s SupplierProductService) BestOffer(ctx context.Context, organizationID uint64, variantID uint64, qty amount.Amount, date time.Time) (*SupplierProduct, error) {
	if err := s.validateVariant(ctx, organizationID, variantID); err != nil {
		return nil, err
	}

	offers, err := s.supplierProducts.ListByItem(ctx, variantID)
	if err != nil {
		return nil, err
	}

	valid := make([]*SupplierProduct, 0, len(offers))
	for _, offer := range offers {
		if offer.Price == nil {
			continue
		}
		if !inDateWindow(date, offer.ValidFrom, offer.ValidTo) {
			continue
		}
		if offer.MinQty > 0 && amount.FromFloat64(offer.MinQty).GreaterThan(qty) {
			continue
		}
		valid = append(valid, offer)
	}
	if len(valid) == 0 {
		return nil, ErrNoValidOffer
	}

	sort.Slice(valid, func(i, j int) bool {
		if valid[i].Priority != valid[j].Priority {
			return valid[i].Priority < valid[j].Priority
		}
		return offerPrice(valid[i]) < offerPrice(valid[j])
	})
	return valid[0], nil
}

func (s SupplierProductService) BestOfferForSupplier(ctx context.Context, variantID, supplierID uint64, qty amount.Amount, date time.Time) (*SupplierProduct, error) {
	offers, err := s.supplierProducts.ListBySupplier(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	valid := make([]*SupplierProduct, 0, len(offers))
	for _, offer := range offers {
		if offer.ItemID != variantID {
			continue
		}
		if offer.Price == nil {
			continue
		}
		if !inDateWindow(date, offer.ValidFrom, offer.ValidTo) {
			continue
		}
		if offer.MinQty > 0 && amount.FromFloat64(offer.MinQty).GreaterThan(qty) {
			continue
		}
		valid = append(valid, offer)
	}
	if len(valid) == 0 {
		return nil, ErrNoValidOffer
	}

	best := valid[0]
	for _, offer := range valid[1:] {
		if offerPrice(offer) < offerPrice(best) {
			best = offer
		}
	}
	return best, nil
}

func (s SupplierProductService) validateVariant(ctx context.Context, organizationID uint64, variantID uint64) error {
	variant, err := s.variants.Find(ctx, variantID)
	if err != nil {
		return err
	}
	if variant == nil {
		return ErrVariantNotFound
	}
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return err
	}
	if template == nil {
		return ErrVariantNotFound
	}
	if template.OrganizationID == nil || *template.OrganizationID != organizationID {
		return ErrVariantNotFound
	}
	return nil
}

func (s SupplierProductService) validateOffer(ctx context.Context, organizationID uint64, offer *SupplierProduct) error {
	if err := s.validateVariant(ctx, organizationID, offer.ItemID); err != nil {
		return err
	}

	contact, err := s.contactDAO.Find(ctx, offer.SupplierID)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrSupplierNotFound
	}
	profile, err := s.supplierDAO.FindByContact(ctx, offer.SupplierID)
	if err != nil {
		return err
	}
	if profile == nil || !profile.Active {
		return ErrSupplierNotSupplier
	}

	if offer.MinQty < 0 {
		return ErrInvalidMinQty
	}
	if offer.ValidFrom != nil && offer.ValidTo != nil && offer.ValidTo.Before(*offer.ValidFrom) {
		return ErrInvalidValidity
	}
	return nil
}

func offerPrice(offer *SupplierProduct) float64 {
	if offer.Price == nil {
		return math.MaxFloat64
	}
	return *offer.Price
}
