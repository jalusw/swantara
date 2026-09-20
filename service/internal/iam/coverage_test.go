package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
)

func TestAuthService_Session(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(session *UserSession, findErr error, updateErr error) AuthService {
		return testAuthActionService(UserDAOMock{}, UserEmailVerificationDAOMock{}, UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{}, UserSessionDAOMock{
			DAOMock: DAOMock[UserSession]{
				FindFunc:   func(_ context.Context, _ uint64) (*UserSession, error) { return session, findErr },
				UpdateFunc: func(_ context.Context, s *UserSession) (*UserSession, error) { return s, updateErr },
			},
			FindByRefreshTokenFunc: func(_ context.Context, _ string) (*UserSession, error) { return session, findErr },
		})
	}

	t.Run("logs out active session", func(t *testing.T) {
		svc := newSvc(&UserSession{Base: model.Base{ID: 1}, UserID: 7}, nil, nil)
		if err := svc.Logout(ctx, "token"); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("logout is idempotent for missing and revoked", func(t *testing.T) {
		svc := newSvc(nil, nil, nil)
		if err := svc.Logout(ctx, "token"); helper.AssertError(t, err, false, nil) {
			return
		}
		now := time.Now()
		svc = newSvc(&UserSession{Base: model.Base{ID: 1}, RevokedAt: &now}, nil, nil)
		if err := svc.Logout(ctx, "token"); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("logout propagates errors", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		helper.AssertError(t, svc.Logout(ctx, "token"), true, dbErr)

		svc = newSvc(&UserSession{Base: model.Base{ID: 1}}, nil, dbErr)
		helper.AssertError(t, svc.Logout(ctx, "token"), true, dbErr)
	})

	t.Run("revokes owned session", func(t *testing.T) {
		svc := newSvc(&UserSession{Base: model.Base{ID: 1}, UserID: 7}, nil, nil)
		if err := svc.RevokeSession(ctx, 7, 1); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("revoke failures", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		helper.AssertError(t, svc.RevokeSession(ctx, 7, 1), true, dbErr)

		svc = newSvc(nil, nil, nil)
		helper.AssertError(t, svc.RevokeSession(ctx, 7, 1), true, ErrSessionNotFound)

		svc = newSvc(&UserSession{Base: model.Base{ID: 1}, UserID: 8}, nil, nil)
		helper.AssertError(t, svc.RevokeSession(ctx, 7, 1), true, ErrSessionNotOwnedByUser)

		svc = newSvc(&UserSession{Base: model.Base{ID: 1}, UserID: 7}, nil, dbErr)
		helper.AssertError(t, svc.RevokeSession(ctx, 7, 1), true, dbErr)
	})
}

func TestMemberService_Count_Provision(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("counts members and propagates error", func(t *testing.T) {
		svc := NewMemberService(MemberDAOMock{
			ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Member, error) {
				return []*Member{{}, {}}, nil
			},
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		got, err := svc.CountMembers(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != 2 {
			t.Errorf("count = %d", got)
		}

		failing := NewMemberService(MemberDAOMock{
			ListByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Member, error) { return nil, dbErr },
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		_, err = failing.CountMembers(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("provision propagates errors", func(t *testing.T) {
		failing := NewMemberService(MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*Member, error) { return nil, dbErr },
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		helper.AssertError(t, failing.ProvisionOwner(ctx, 1, 10), true, dbErr)

		roleFailing := NewMemberService(MemberDAOMock{}, MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) { return nil, dbErr },
		}, PermissionDAOMock{})
		helper.AssertError(t, roleFailing.ProvisionOwner(ctx, 1, 10), true, dbErr)
	})
}

func TestMemberService_Invite_CreateRole_Remove(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("invite propagates errors", func(t *testing.T) {
		failing := NewMemberService(MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*Member, error) { return nil, dbErr },
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		_, err := failing.Invite(ctx, 1, 20)
		helper.AssertError(t, err, true, dbErr)

		noRole := NewMemberService(MemberDAOMock{}, MemberRoleDAOMock{}, PermissionDAOMock{})
		_, err = noRole.Invite(ctx, 1, 20)
		helper.AssertError(t, err, true, ErrMemberRoleNotFound)
	})

	t.Run("create role propagates errors", func(t *testing.T) {
		failing := NewMemberService(MemberDAOMock{}, MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) { return nil, dbErr },
		}, PermissionDAOMock{})
		_, err := failing.CreateRole(ctx, 1, &MemberRole{Code: "x"}, nil)
		helper.AssertError(t, err, true, dbErr)

		exists := NewMemberService(MemberDAOMock{}, MemberRoleDAOMock{
			FindByOrganizationAndCodeFunc: func(_ context.Context, _ uint64, _ string) (*MemberRole, error) {
				return &MemberRole{Base: model.Base{ID: 3}}, nil
			},
		}, PermissionDAOMock{})
		_, err = exists.CreateRole(ctx, 1, &MemberRole{Code: "x"}, nil)
		helper.AssertError(t, err, true, ErrMemberRoleExists)
	})

	t.Run("remove propagates errors", func(t *testing.T) {
		failing := NewMemberService(MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				FindFunc: func(_ context.Context, _ uint64) (*Member, error) { return nil, dbErr },
			},
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		helper.AssertError(t, failing.RemoveMember(ctx, 1, 2), true, dbErr)

		foreign := NewMemberService(MemberDAOMock{
			CRUDMock: dao.CRUDMock[Member]{
				FindFunc: func(_ context.Context, _ uint64) (*Member, error) {
					return &Member{Base: model.Base{ID: 2}, OrganizationID: 99}, nil
				},
			},
		}, MemberRoleDAOMock{}, PermissionDAOMock{})
		helper.AssertError(t, foreign.RemoveMember(ctx, 1, 2), true, ErrMemberNotFound)
	})
}

func TestIAMFixtures_Tables(t *testing.T) {
	if UserFixture() == nil {
		t.Error("user = nil")
	}
	if UserSessionFixture() == nil {
		t.Error("session = nil")
	}
	if UserFixture(func(u *User) *User { return u }) == nil {
		t.Error("user opt = nil")
	}
}

func TestIAMMock_Fallbacks(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("user dao", func(t *testing.T) {
		bare := UserDAOMock{}
		if _, err := bare.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.FindByEmail(ctx, "a@b.c"); err != nil {
			t.Errorf("FindByEmail = %v", err)
		}
		if _, err := bare.FindByPhone(ctx, "1"); err != nil {
			t.Errorf("FindByPhone = %v", err)
		}
		wired := UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*User, error) { return &User{}, nil },
			FindByPhoneFunc: func(_ context.Context, _ string) (*User, error) { return &User{}, nil },
		}
		if _, err := wired.FindByEmail(ctx, "a@b.c"); err != nil {
			t.Errorf("FindByEmail = %v", err)
		}
		if _, err := wired.FindByPhone(ctx, "1"); err != nil {
			t.Errorf("FindByPhone = %v", err)
		}
	})

	t.Run("session dao", func(t *testing.T) {
		bare := UserSessionDAOMock{}
		if _, err := bare.FindByRefreshToken(ctx, "t"); err != nil {
			t.Errorf("FindByRefreshToken = %v", err)
		}
		if _, err := bare.ListUserSessions(ctx, 1); err != nil {
			t.Errorf("ListUserSessions = %v", err)
		}
		if err := bare.RevokeByUserID(ctx, 1); err != nil {
			t.Errorf("RevokeByUserID = %v", err)
		}
		wired := UserSessionDAOMock{
			FindByRefreshTokenFunc: func(_ context.Context, _ string) (*UserSession, error) { return &UserSession{}, nil },
			RevokeByUserIDFunc:     func(_ context.Context, _ uint64) error { return nil },
		}
		if _, err := wired.FindByRefreshToken(ctx, "t"); err != nil {
			t.Errorf("FindByRefreshToken = %v", err)
		}
		if err := wired.RevokeByUserID(ctx, 1); err != nil {
			t.Errorf("RevokeByUserID = %v", err)
		}
	})

	t.Run("permission dao", func(t *testing.T) {
		bare := PermissionDAOMock{}
		if _, err := bare.FindByCode(ctx, "c"); err != nil {
			t.Errorf("FindByCode = %v", err)
		}
		if _, err := bare.UserIsOrganizationOwner(ctx, 1, 2); err != nil {
			t.Errorf("UserIsOrganizationOwner = %v", err)
		}
		if _, err := bare.ListUserOrganizationPermissions(ctx, 1, 2); err != nil {
			t.Errorf("ListUserOrganizationPermissions = %v", err)
		}
		if _, err := bare.UserHasOrganizationPermission(ctx, 1, 2, "r", "a"); err != nil {
			t.Errorf("UserHasOrganizationPermission = %v", err)
		}
		wired := PermissionDAOMock{
			UserIsOrganizationOwnerFunc: func(_ context.Context, _, _ uint64) (bool, error) { return true, nil },
			ListUserOrganizationPermissionsFunc: func(_ context.Context, _, _ uint64) ([]*Permission, error) {
				return []*Permission{{}}, nil
			},
		}
		if _, err := wired.UserIsOrganizationOwner(ctx, 1, 2); err != nil {
			t.Errorf("UserIsOrganizationOwner = %v", err)
		}
		if _, err := wired.ListUserOrganizationPermissions(ctx, 1, 2); err != nil {
			t.Errorf("ListUserOrganizationPermissions = %v", err)
		}
	})

	t.Run("member dao lists", func(t *testing.T) {
		bare := MemberDAOMock{}
		if _, err := bare.ListOrganizationsByUser(ctx, 1); err != nil {
			t.Errorf("ListOrganizationsByUser = %v", err)
		}
		if _, err := bare.ListByOrganization(ctx, 1); err != nil {
			t.Errorf("ListByOrganization = %v", err)
		}
	})
}
