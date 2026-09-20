package organization

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestOrganizationService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		org      *reference.Organization
		wantErrV error
	}{
		{
			name: "creates organization with base currency",
			org:  &reference.Organization{Name: "Acme", BaseCurrency: "IDR"},
		},
		{
			name:     "rejects empty name",
			org:      &reference.Organization{Name: "", BaseCurrency: "IDR"},
			wantErrV: ErrNameRequired,
		},
		{
			name:     "rejects unknown base currency",
			org:      &reference.Organization{Name: "Acme", BaseCurrency: "XYZ"},
			wantErrV: ErrBaseCurrencyNotFound,
		},
		{
			name:     "rejects missing parent",
			org:      &reference.Organization{Name: "Acme", BaseCurrency: "IDR", ParentID: uint64Ptr(99)},
			wantErrV: ErrParentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgs := DAOMock{
				CRUDMock: dao.CRUDMock[reference.Organization]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Organization, error) {
						return nil, nil
					},
				},
			}
			currencies := currencyDAOMock{
				searchFunc: func(_ context.Context, _ string, value any) (*reference.Currency, error) {
					if value == "XYZ" {
						return nil, nil
					}
					return &reference.Currency{Code: "IDR"}, nil
				},
			}
			svc := NewOrganizationService(orgs, currencies, memberProvisionerMock{})

			created, err := svc.Create(ctx, tt.org, 1)
			if helper.AssertError(t, err, tt.wantErrV != nil, tt.wantErrV) {
				return
			}
			if created == nil {
				t.Fatal("expected created organization, got nil")
			}
			if created.Timezone == "" {
				t.Error("expected default timezone to be applied")
			}
			if created.TaxYearStartMonth == 0 {
				t.Error("expected default tax year start month to be applied")
			}
		})
	}
}

func TestOrganizationService_Create_ProvisionsOwner(t *testing.T) {
	ctx := context.Background()

	var provisionedOrgID uint64
	var provisionedUserID uint64
	members := memberProvisionerMock{
		provisionOwnerFunc: func(_ context.Context, organizationID, ownerUserID uint64) error {
			provisionedOrgID = organizationID
			provisionedUserID = ownerUserID
			return nil
		},
	}

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			CreateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
				org.ID = 7
				return org, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Organization, error) {
				return nil, nil
			},
		},
	}
	currencies := currencyDAOMock{
		searchFunc: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}
	svc := NewOrganizationService(orgs, currencies, members)

	created, err := svc.Create(ctx, &reference.Organization{Name: "Acme", BaseCurrency: "IDR"}, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 7 {
		t.Errorf("expected created organization id 7, got %d", created.ID)
	}
	if provisionedOrgID != 7 {
		t.Errorf("expected owner provisioned for org 7, got %d", provisionedOrgID)
	}
	if provisionedUserID != 42 {
		t.Errorf("expected owner provisioned for user 42, got %d", provisionedUserID)
	}
}

func TestOrganizationService_Update_RejectsCycle(t *testing.T) {
	ctx := context.Background()
	org := &reference.Organization{Base: model.Base{ID: 3}, Name: "Child"}

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
				switch id {
				case 3:
					return org, nil
				case 1:
					return &reference.Organization{Base: model.Base{ID: 1}, Name: "Root", ParentID: uint64Ptr(3)}, nil
				default:
					return nil, nil
				}
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 3, &reference.Organization{
		Name:     "Child",
		ParentID: uint64Ptr(1),
	})
	if !helper.AssertError(t, err, true, ErrParentCycle) {
		return
	}
}

func TestOrganizationService_Update_NotFound(t *testing.T) {
	ctx := context.Background()

	orgs := DAOMock{
		CRUDMock: dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return nil, nil
			},
		},
	}
	svc := NewOrganizationService(orgs, currencyDAOMock{}, memberProvisionerMock{})

	_, err := svc.Update(ctx, 1, &reference.Organization{Name: "Acme"})
	if helper.AssertError(t, err, true, ErrOrganizationNotFound) {
		return
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

type currencyDAOMock struct {
	searchFunc func(ctx context.Context, field string, value any) (*reference.Currency, error)
}

func (m currencyDAOMock) Search(ctx context.Context, field string, value any) (*reference.Currency, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, field, value)
	}
	return nil, nil
}

type memberProvisionerMock struct {
	provisionOwnerFunc func(ctx context.Context, organizationID, ownerUserID uint64) error
}

func (m memberProvisionerMock) ProvisionOwner(ctx context.Context, organizationID, ownerUserID uint64) error {
	if m.provisionOwnerFunc != nil {
		return m.provisionOwnerFunc(ctx, organizationID, ownerUserID)
	}
	return nil
}
