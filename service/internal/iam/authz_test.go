package iam

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestNewAuthzService(t *testing.T) {
	svc := NewAuthzService(PermissionDAOMock{})
	if svc.permissionDAO == nil {
		t.Error("expected non-nil permissionDAO")
	}
}

func permissionCodes(permissions []*Permission) []string {
	codes := make([]string, len(permissions))
	for i, permission := range permissions {
		codes[i] = permission.Code
	}
	return codes
}

func TestEffectiveOrganizationPermissions(t *testing.T) {
	ctx := context.Background()

	customerView := PermissionFixture(func(p *Permission) *Permission { p.Code = "customer.view"; return p })
	customerCreate := PermissionFixture(func(p *Permission) *Permission { p.Code = "customer.create"; return p })
	organizationView := PermissionFixture(func(p *Permission) *Permission { p.Code = "organization.view"; return p })
	catalog := []*Permission{customerView, customerCreate, organizationView}
	orgPerms := []*Permission{customerView, customerCreate}

	tests := []struct {
		name         string
		ownerFunc    func(ctx context.Context, userID, organizationID uint64) (bool, error)
		listFunc     func(ctx context.Context, q *query.Query) (*query.Page[Permission], error)
		orgPermsFunc func(ctx context.Context, userID, organizationID uint64) ([]*Permission, error)
		wantCodes    []string
		wantErr      bool
	}{
		{
			name: "owner receives full catalog",
			ownerFunc: func(ctx context.Context, userID, organizationID uint64) (bool, error) {
				return true, nil
			},
			listFunc: func(ctx context.Context, q *query.Query) (*query.Page[Permission], error) {
				return &query.Page[Permission]{Items: catalog, Count: int64(len(catalog))}, nil
			},
			wantCodes: []string{"customer.view", "customer.create", "organization.view"},
		},
		{
			name: "non-owner receives org permissions",
			ownerFunc: func(ctx context.Context, userID, organizationID uint64) (bool, error) {
				return false, nil
			},
			orgPermsFunc: func(ctx context.Context, userID, organizationID uint64) ([]*Permission, error) {
				return orgPerms, nil
			},
			wantCodes: []string{"customer.view", "customer.create"},
		},
		{
			name: "owner check error",
			ownerFunc: func(ctx context.Context, userID, organizationID uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "list error for owner",
			ownerFunc: func(ctx context.Context, userID, organizationID uint64) (bool, error) {
				return true, nil
			},
			listFunc: func(ctx context.Context, q *query.Query) (*query.Page[Permission], error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "org permissions error for non-owner",
			ownerFunc: func(ctx context.Context, userID, organizationID uint64) (bool, error) {
				return false, nil
			},
			orgPermsFunc: func(ctx context.Context, userID, organizationID uint64) ([]*Permission, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthzService(
				PermissionDAOMock{
					DAOMock:                             DAOMock[Permission]{ListFunc: tt.listFunc},
					UserIsOrganizationOwnerFunc:         tt.ownerFunc,
					ListUserOrganizationPermissionsFunc: tt.orgPermsFunc,
				},
			)
			permissions, err := svc.EffectiveOrganizationPermissions(ctx, 7, 3)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(tt.wantCodes, permissionCodes(permissions)) {
				t.Errorf("permission codes = %v, want %v", permissionCodes(permissions), tt.wantCodes)
			}
		})
	}
}
