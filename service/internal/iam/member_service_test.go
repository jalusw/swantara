package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestMemberService_ProvisionOwner(t *testing.T) {
	ctx := context.Background()

	t.Run("creates owner role, member role, member and assignment", func(t *testing.T) {
		var assignedRoleID uint64
		members := MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				CreateFunc: func(_ context.Context, member *Member) (*Member, error) {
					member.ID = 9
					return member, nil
				},
			},
			AssignRoleFunc: func(_ context.Context, memberID, roleID uint64) error {
				assignedRoleID = roleID
				return nil
			},
		}
		roles := MemberRoleDAOMock{
			CRUDMock: dao.CRUDMock[MemberRole]{
				CreateFunc: func(_ context.Context, role *MemberRole) (*MemberRole, error) {
					role.ID = 5
					return role, nil
				},
			},
		}
		permissions := PermissionDAOMock{
			DAOMock: DAOMock[Permission]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Permission], error) {
					return &query.Page[Permission]{Items: []*Permission{{Base: model.Base{ID: 1}, Action: "view"}, {Base: model.Base{ID: 2}, Action: "create"}}}, nil
				},
			},
		}
		svc := NewMemberService(members, roles, permissions)

		err := svc.ProvisionOwner(ctx, 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if assignedRoleID != 5 {
			t.Errorf("expected owner role id 5 to be assigned, got %d", assignedRoleID)
		}
	})

	t.Run("returns nil when owner membership already exists", func(t *testing.T) {
		members := MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*Member, error) {
				return &Member{Base: model.Base{ID: 3}}, nil
			},
		}
		roles := MemberRoleDAOMock{}
		permissions := PermissionDAOMock{}
		svc := NewMemberService(members, roles, permissions)

		if err := svc.ProvisionOwner(ctx, 1, 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMemberService_Invite(t *testing.T) {
	ctx := context.Background()

	t.Run("creates member with default member role", func(t *testing.T) {
		var assignedRoleID uint64
		members := MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				CreateFunc: func(_ context.Context, member *Member) (*Member, error) {
					member.ID = 12
					return member, nil
				},
			},
			AssignRoleFunc: func(_ context.Context, _, roleID uint64) error {
				assignedRoleID = roleID
				return nil
			},
		}
		roles := MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, code string) (*MemberRole, error) {
				return &MemberRole{Base: model.Base{ID: 6}, Code: code}, nil
			},
		}
		permissions := PermissionDAOMock{}
		svc := NewMemberService(members, roles, permissions)

		member, err := svc.Invite(ctx, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if member.ID != 12 {
			t.Errorf("expected member id 12, got %d", member.ID)
		}
		if assignedRoleID != 6 {
			t.Errorf("expected member role id 6 to be assigned, got %d", assignedRoleID)
		}
	})

	t.Run("rejects when user is already a member", func(t *testing.T) {
		members := MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*Member, error) {
				return &Member{Base: model.Base{ID: 3}}, nil
			},
		}
		roles := MemberRoleDAOMock{}
		permissions := PermissionDAOMock{}
		svc := NewMemberService(members, roles, permissions)

		_, err := svc.Invite(ctx, 1, 20)
		if !helper.AssertError(t, err, true, ErrAlreadyMember) {
			return
		}
	})

	t.Run("rejects when default member role is missing", func(t *testing.T) {
		members := MemberDAOMock{}
		roles := MemberRoleDAOMock{}
		permissions := PermissionDAOMock{}
		svc := NewMemberService(members, roles, permissions)

		_, err := svc.Invite(ctx, 1, 20)
		if !helper.AssertError(t, err, true, ErrMemberRoleNotFound) {
			return
		}
	})
}

func TestMemberService_CreateRole(t *testing.T) {
	ctx := context.Background()

	t.Run("creates role and binds permissions", func(t *testing.T) {
		var boundRoleID uint64
		roles := MemberRoleDAOMock{
			CRUDMock: dao.CRUDMock[MemberRole]{
				CreateFunc: func(_ context.Context, role *MemberRole) (*MemberRole, error) {
					role.ID = 8
					return role, nil
				},
			},
			SetPermissionsFunc: func(_ context.Context, roleID uint64, permissionIDs []uint64) error {
				boundRoleID = roleID
				return nil
			},
		}
		svc := NewMemberService(MemberDAOMock{}, roles, PermissionDAOMock{})

		role, err := svc.CreateRole(ctx, 1, &MemberRole{Name: "Accountant", Code: "accountant"}, []uint64{1, 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if role.ID != 8 {
			t.Errorf("expected role id 8, got %d", role.ID)
		}
		if boundRoleID != 8 {
			t.Errorf("expected permissions bound to role id 8, got %d", boundRoleID)
		}
		if role.OrganizationID != 1 {
			t.Errorf("expected role organization id 1, got %d", role.OrganizationID)
		}
	})

	t.Run("rejects duplicate role code", func(t *testing.T) {
		roles := MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) {
				return &MemberRole{Base: model.Base{ID: 8}}, nil
			},
		}
		svc := NewMemberService(MemberDAOMock{}, roles, PermissionDAOMock{})

		_, err := svc.CreateRole(ctx, 1, &MemberRole{Name: "Accountant", Code: "accountant"}, nil)
		if !errors.Is(err, ErrMemberRoleExists) {
			t.Errorf("expected ErrMemberRoleExists, got %v", err)
		}
	})
}

func TestMemberService_RemoveMember(t *testing.T) {
	ctx := context.Background()

	t.Run("removes member belonging to organization", func(t *testing.T) {
		var deletedID uint64
		members := MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				FindFunc: func(_ context.Context, _ uint64) (*Member, error) {
					return &Member{Base: model.Base{ID: 4}, OrganizationID: 1}, nil
				},
				HardDeleteFunc: func(_ context.Context, id uint64) error {
					deletedID = id
					return nil
				},
			},
		}
		svc := NewMemberService(members, MemberRoleDAOMock{}, PermissionDAOMock{})

		if err := svc.RemoveMember(ctx, 1, 4); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deletedID != 4 {
			t.Errorf("expected member 4 deleted, got %d", deletedID)
		}
	})

	t.Run("rejects member that does not belong to organization", func(t *testing.T) {
		members := MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				FindFunc: func(_ context.Context, _ uint64) (*Member, error) {
					return &Member{Base: model.Base{ID: 4}, OrganizationID: 99}, nil
				},
			},
		}
		svc := NewMemberService(members, MemberRoleDAOMock{}, PermissionDAOMock{})

		err := svc.RemoveMember(ctx, 1, 4)
		if !helper.AssertError(t, err, true, ErrMemberNotFound) {
			return
		}
	})

	t.Run("rejects missing member", func(t *testing.T) {
		members := MemberDAOMock{}
		svc := NewMemberService(members, MemberRoleDAOMock{}, PermissionDAOMock{})

		err := svc.RemoveMember(ctx, 1, 4)
		if !helper.AssertError(t, err, true, ErrMemberNotFound) {
			return
		}
	})
}
