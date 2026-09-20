package products

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestSupplierProductService_BestOffer(t *testing.T) {
	ctx := context.Background()
	orgID := uint64(7)
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	validFrom := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		variants ItemVariantDAOMock
		offers   SupplierProductDAOMock
		wantID   uint64
		wantErr  error
	}{
		{
			name:     "picks highest priority valid offer",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(80.0), Priority: 10},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 11, MinQty: 1, Price: ptr(85.0), Priority: 5},
				{Base: model.Base{ID: 3}, ItemID: 100, SupplierID: 12, MinQty: 1, Price: ptr(60.0), Priority: 10},
			}),
			wantID: 2,
		},
		{
			name:     "tie breaks by price",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(80.0), Priority: 10},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 11, MinQty: 1, Price: ptr(70.0), Priority: 10},
			}),
			wantID: 2,
		},
		{
			name:     "honors validity window",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(50.0), Priority: 1, ValidFrom: &validFrom, ValidTo: &validTo},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 11, MinQty: 1, Price: ptr(90.0), Priority: 10},
			}),
			wantID: 2,
		},
		{
			name:     "honors min qty",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 50, Price: ptr(40.0), Priority: 1},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 11, MinQty: 1, Price: ptr(95.0), Priority: 10},
			}),
			wantID: 2,
		},
		{
			name:     "returns no valid offer when outside validity",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(50.0), Priority: 1, ValidTo: &validTo},
			}),
			wantErr: ErrNoValidOffer,
		},
		{
			name:     "rejects priceless offer",
			variants: foundVariantMock(100),
			offers: offersMock([]*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Priority: 1},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 10, MinQty: 1, Priority: 2},
			}),
			wantErr: ErrNoValidOffer,
		},
		{
			name:    "rejects unknown variant",
			offers:  SupplierProductDAOMock{},
			wantErr: ErrVariantNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSupplierProductService(tt.variants, foundTemplateMock(orgID), tt.offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

			best, err := svc.BestOffer(ctx, orgID, 100, amount.FromFloat64(10), date)

			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if best.ID != tt.wantID {
				t.Errorf("best offer = %d, want %d", best.ID, tt.wantID)
			}
		})
	}
}

func TestSupplierProductService_Create_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	validFrom := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
		}),
	}
	contactFound := contacts.ContactDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return &contacts.Contact{Base: model.Base{ID: 10}, Name: "Supplier"}, nil
		}),
	}
	activeSupplier := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return &contacts.SupplierProfile{ContactID: 10, Active: true}, nil
		},
	}
	inactiveSupplier := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return &contacts.SupplierProfile{ContactID: 10, Active: false}, nil
		},
	}

	tests := []struct {
		name      string
		variants  ItemVariantDAOMock
		contacts  contacts.ContactDAOMock
		suppliers contacts.SupplierProfileDAOMock
		item      *SupplierProduct
		wantErr   error
	}{
		{
			name:    "rejects unknown variant",
			item:    &SupplierProduct{ItemID: 100, SupplierID: 10},
			wantErr: ErrVariantNotFound,
		},
		{
			name:      "rejects non supplier supplier",
			variants:  variants,
			contacts:  contactFound,
			suppliers: inactiveSupplier,
			item:      &SupplierProduct{ItemID: 100, SupplierID: 10},
			wantErr:   ErrSupplierNotSupplier,
		},
		{
			name:      "rejects negative min qty",
			variants:  variants,
			contacts:  contactFound,
			suppliers: activeSupplier,
			item:      &SupplierProduct{ItemID: 100, SupplierID: 10, MinQty: -1},
			wantErr:   ErrInvalidMinQty,
		},
		{
			name:      "rejects invalid validity",
			variants:  variants,
			contacts:  contactFound,
			suppliers: activeSupplier,
			item:      &SupplierProduct{ItemID: 100, SupplierID: 10, ValidFrom: &validFrom, ValidTo: &validTo},
			wantErr:   ErrInvalidValidity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSupplierProductService(tt.variants, foundTemplateMock(7), SupplierProductDAOMock{}, tt.contacts, tt.suppliers)

			_, err := svc.Create(ctx, 7, tt.item)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func foundVariantMock(id uint64) ItemVariantDAOMock {
	return ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: id}, ItemID: 1}, nil
		}),
	}
}

func foundTemplateMock(orgID uint64) ItemDAOMock {
	return ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}, OrganizationID: &orgID}, nil
		}),
	}
}

func offersMock(offers []*SupplierProduct) SupplierProductDAOMock {
	return SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return offers, nil
		},
	}
}

func TestSupplierProductService_BestOfferForSupplier_FiltersByVendorAndProduct(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	offers := SupplierProductDAOMock{
		ListBySupplierFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return []*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(80.0), Priority: 10},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(60.0), Priority: 10},
				{Base: model.Base{ID: 3}, ItemID: 200, SupplierID: 10, MinQty: 1, Price: ptr(50.0), Priority: 10},
			}, nil
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	best, err := svc.BestOfferForSupplier(ctx, 100, 10, amount.FromFloat64(10), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if best.ID != 2 {
		t.Errorf("best offer = %d, want 2", best.ID)
	}
}

func TestSupplierProductService_BestOfferForSupplier_ReturnsNoOfferWhenVendorUnrelated(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	offers := SupplierProductDAOMock{
		ListBySupplierFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return []*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 200, SupplierID: 10, MinQty: 1, Price: ptr(80.0), Priority: 10},
			}, nil
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOfferForSupplier(ctx, 100, 10, amount.FromFloat64(10), date)
	if helper.AssertError(t, err, true, ErrNoValidOffer) {
		return
	}
}

func TestSupplierProductService_Create_SetsDefaultsAndCreates(t *testing.T) {
	ctx := context.Background()

	offers := SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[SupplierProduct]{
			CreateFunc: func(_ context.Context, offer *SupplierProduct) (*SupplierProduct, error) {
				offer.ID = 1
				return offer, nil
			},
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contactFoundMock(), activeSupplierMock())

	created, err := svc.Create(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || created.MinQty != 1 || created.Priority != DefaultSupplierPriority {
		t.Errorf("created = %+v, want id 1 with default min qty and priority", created)
	}
}

func TestSupplierProductService_Create_KeepsExplicitDefaults(t *testing.T) {
	ctx := context.Background()

	offers := SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[SupplierProduct]{
			CreateFunc: func(_ context.Context, offer *SupplierProduct) (*SupplierProduct, error) {
				return offer, nil
			},
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contactFoundMock(), activeSupplierMock())

	created, err := svc.Create(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10, MinQty: 5, Priority: 3})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.MinQty != 5 || created.Priority != 3 {
		t.Errorf("created = %+v, want min qty 5 and priority 3", created)
	}
}

func TestSupplierProductService_Update_UpdatesOffer(t *testing.T) {
	ctx := context.Background()

	offers := SupplierProductDAOMock{
		CRUDMock: dao.CRUDMock[SupplierProduct]{
			UpdateFunc: func(_ context.Context, offer *SupplierProduct) (*SupplierProduct, error) {
				return offer, nil
			},
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contactFoundMock(), activeSupplierMock())

	updated, err := svc.Update(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10, MinQty: 2, Price: ptr(75.0)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.MinQty != 2 || updated.Price == nil || *updated.Price != 75.0 {
		t.Errorf("updated = %+v, want min qty 2 and price 75", updated)
	}
}

func TestSupplierProductService_Update_RejectsInvalidOffer(t *testing.T) {
	ctx := context.Background()

	svc := NewSupplierProductService(ItemVariantDAOMock{}, foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.Update(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10})

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestSupplierProductService_BestOffer_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	offers := SupplierProductDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 7, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_BestOfferForSupplier_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	offers := SupplierProductDAOMock{
		ListBySupplierFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOfferForSupplier(ctx, 100, 10, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_BestOffer_PropagatesVariantError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewSupplierProductService(variants, foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 7, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_BestOffer_PropagatesTemplateError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewSupplierProductService(foundVariantMock(100), templates, SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 7, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_BestOffer_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewSupplierProductService(foundVariantMock(100), templates, SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 7, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestSupplierProductService_BestOffer_RejectsVariantFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 8, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestSupplierProductService_Create_PropagatesContactLookupError(t *testing.T) {
	ctx := context.Background()

	contactDAO := contacts.ContactDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contactDAO, contacts.SupplierProfileDAOMock{})

	_, err := svc.Create(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_Create_RejectsUnknownVendor(t *testing.T) {
	ctx := context.Background()

	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.Create(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10})

	if helper.AssertError(t, err, true, ErrSupplierNotFound) {
		return
	}
}

func TestSupplierProductService_Create_PropagatesSupplierLookupError(t *testing.T) {
	ctx := context.Background()

	suppliers := contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contactFoundMock(), suppliers)

	_, err := svc.Create(ctx, 7, &SupplierProduct{ItemID: 100, SupplierID: 10})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSupplierProductService_BestOffer_ReturnsNoOfferWhenEmpty(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOffer(ctx, 7, 100, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, ErrNoValidOffer) {
		return
	}
}

func TestSupplierProductService_BestOfferForSupplier_ReturnsNoOfferWhenEmpty(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), SupplierProductDAOMock{}, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	_, err := svc.BestOfferForSupplier(ctx, 100, 10, amount.FromFloat64(10), date)

	if helper.AssertError(t, err, true, ErrNoValidOffer) {
		return
	}
}

func TestSupplierProductService_BestOfferForSupplier_FiltersInvalidOffers(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	offers := SupplierProductDAOMock{
		ListBySupplierFunc: func(_ context.Context, _ uint64) ([]*SupplierProduct, error) {
			return []*SupplierProduct{
				{Base: model.Base{ID: 1}, ItemID: 100, SupplierID: 10, MinQty: 1, Priority: 1},
				{Base: model.Base{ID: 2}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(70.0), Priority: 10, ValidTo: &validTo},
				{Base: model.Base{ID: 3}, ItemID: 100, SupplierID: 10, MinQty: 50, Price: ptr(40.0), Priority: 10},
				{Base: model.Base{ID: 4}, ItemID: 100, SupplierID: 10, MinQty: 1, Price: ptr(60.0), Priority: 10},
			}, nil
		},
	}
	svc := NewSupplierProductService(foundVariantMock(100), foundTemplateMock(7), offers, contacts.ContactDAOMock{}, contacts.SupplierProfileDAOMock{})

	best, err := svc.BestOfferForSupplier(ctx, 100, 10, amount.FromFloat64(10), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if best.ID != 4 {
		t.Errorf("best offer = %d, want 4 (only valid priced offer)", best.ID)
	}
}

func TestOfferPrice_ReturnsMaxFloatForNilPrice(t *testing.T) {
	if offerPrice(&SupplierProduct{}) != math.MaxFloat64 {
		t.Error("offerPrice = not MaxFloat64 for priceless offer")
	}
	if offerPrice(&SupplierProduct{Price: ptr(50.0)}) != 50.0 {
		t.Error("offerPrice = not 50.0 for priced offer")
	}
}

func contactFoundMock() contacts.ContactDAOMock {
	return contacts.ContactDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return &contacts.Contact{Base: model.Base{ID: 10}, Name: "Supplier"}, nil
		}),
	}
}

func activeSupplierMock() contacts.SupplierProfileDAOMock {
	return contacts.SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
			return &contacts.SupplierProfile{ContactID: 10, Active: true}, nil
		},
	}
}
