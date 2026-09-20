package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestOrganizationService_Update_AppliesFields(t *testing.T) {
	ctx := context.Background()
	existing := &reference.Organization{Base: model.Base{ID: 3}, Name: "Old", BaseCurrency: "IDR"}
	legalName := "New Legal"
	countryCode := "US"
	taxID := "TAX1"

	var updated *reference.Organization
	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return existing, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Organization, error) {
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
				updated = org
				return org, nil
			},
		},
	}
	currencies := currencyDAOMock{
		searchFunc: func(_ context.Context, _ string, value any) (*reference.Currency, error) {
			if value == "USD" {
				return &reference.Currency{Code: "USD"}, nil
			}
			return nil, nil
		},
	}
	svc := NewOrganizationService(orgs, currencies, memberProvisionerMock{})

	result, err := svc.Update(ctx, 3, &reference.Organization{
		Name:              "New",
		LegalName:         &legalName,
		BaseCurrency:      "USD",
		ParentID:          uint64Ptr(1),
		CountryCode:       &countryCode,
		TaxID:             &taxID,
		Timezone:          "America/New_York",
		TaxYearStartMonth: 7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated.Name != "New" || updated.BaseCurrency != "USD" {
		t.Errorf("updated = %+v, want applied fields", updated)
	}
	if result.Name != "New" || result.Timezone != "America/New_York" || result.TaxYearStartMonth != 7 {
		t.Errorf("result = %+v, want applied optional fields", result)
	}
}

func TestOrganizationService_Update_RejectsEmptyName(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return &reference.Organization{Base: model.Base{ID: 3}}, nil
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "  "})
	if !helper.AssertError(t, err, true, ErrNameRequired) {
		return
	}
}

func TestOrganizationService_Update_RejectsUnknownCurrency(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return &reference.Organization{Base: model.Base{ID: 3}}, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Organization, error) {
				return nil, nil
			},
		},
	}
	currencies := currencyDAOMock{
		searchFunc: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return nil, nil
		},
	}
	svc := NewOrganizationService(orgs, currencies, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", BaseCurrency: "XYZ"})
	if !helper.AssertError(t, err, true, ErrBaseCurrencyNotFound) {
		return
	}
}

func TestOrganizationService_Update_SelfParentCycle(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				if id == 3 {
					return &reference.Organization{Base: model.Base{ID: 3}}, nil
				}
				return nil, nil
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", ParentID: uint64Ptr(3)})
	if !helper.AssertError(t, err, true, ErrParentCycle) {
		return
	}
}

func TestOrganizationService_Update_RejectsMissingParent(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				if id == 3 {
					return &reference.Organization{Base: model.Base{ID: 3}}, nil
				}
				return nil, nil
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", ParentID: uint64Ptr(99)})
	if !helper.AssertError(t, err, true, ErrParentNotFound) {
		return
	}
}

func TestOrganizationService_Create_RejectsSelfParent(t *testing.T) {
	ctx := context.Background()
	currencies := currencyDAOMock{
		searchFunc: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}

	svc := NewOrganizationService(DAOMock{}, currencies, memberProvisionerMock{})

	_, err := svc.Create(ctx, &reference.Organization{Name: "Acme", BaseCurrency: "IDR", ParentID: uint64Ptr(0)}, 1)
	if !helper.AssertError(t, err, true, ErrParentCycle) {
		return
	}
}

func TestOrganizationService_Create_PropagatesCurrencyLookupError(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	currencies := currencyDAOMock{
		searchFunc: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return nil, dbErr
		},
	}
	svc := NewOrganizationService(DAOMock{}, currencies, memberProvisionerMock{})

	_, err := svc.Create(ctx, &reference.Organization{Name: "Acme", BaseCurrency: "IDR"}, 1)
	if !helper.AssertError(t, err, true, dbErr) {
		return
	}
}

func TestOrganizationService_Update_PropagatesParentLookupError(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				if id == 3 {
					return &reference.Organization{Base: model.Base{ID: 3}}, nil
				}
				return nil, dbErr
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", ParentID: uint64Ptr(9)})
	if !helper.AssertError(t, err, true, dbErr) {
		return
	}
}

func TestOrganizationService_Update_DeepParentChain(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				switch id {
				case 3:
					return &reference.Organization{Base: model.Base{ID: 3}}, nil
				case 9:
					return &reference.Organization{Base: model.Base{ID: 9}, ParentID: uint64Ptr(5)}, nil
				case 5:
					return nil, nil
				default:
					return nil, nil
				}
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", ParentID: uint64Ptr(9)})
	if !helper.AssertError(t, err, true, ErrParentNotFound) {
		return
	}
}

func TestOrganizationService_Update_DeepParentChainCycle(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				switch id {
				case 3:
					return &reference.Organization{Base: model.Base{ID: 3}}, nil
				case 9:
					return &reference.Organization{Base: model.Base{ID: 9}, ParentID: uint64Ptr(3)}, nil
				default:
					return nil, nil
				}
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{Name: "Acme", ParentID: uint64Ptr(9)})
	if !helper.AssertError(t, err, true, ErrParentCycle) {
		return
	}
}
