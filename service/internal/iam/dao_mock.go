package iam

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type DAOMock[E any] struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[E], error)
	FindFunc       func(ctx context.Context, id uint64) (*E, error)
	CreateFunc     func(ctx context.Context, entity *E) (*E, error)
	UpdateFunc     func(ctx context.Context, entity *E) (*E, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	SearchFunc     func(ctx context.Context, field string, value any) (*E, error)
}

func (m DAOMock[E]) Search(ctx context.Context, field string, value any) (*E, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return nil, nil
}

func (m DAOMock[E]) List(ctx context.Context, q *query.Query) (*query.Page[E], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[E]{Items: []*E{}, Count: 0}, nil
}

func (m DAOMock[E]) Find(ctx context.Context, id uint64) (*E, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m DAOMock[E]) Create(ctx context.Context, entity *E) (*E, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m DAOMock[E]) Update(ctx context.Context, entity *E) (*E, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m DAOMock[E]) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m DAOMock[E]) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

type UserDAOMock struct {
	DAOMock[User]
	FindByEmailFunc    func(ctx context.Context, email string) (*User, error)
	FindByUsernameFunc func(ctx context.Context, username string) (*User, error)
	FindByPhoneFunc    func(ctx context.Context, phone string) (*User, error)
	SearchUsersFunc    func(ctx context.Context, term string, limit int) ([]*User, error)
}

func (m UserDAOMock) FindByEmail(ctx context.Context, email string) (*User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m UserDAOMock) FindByUsername(ctx context.Context, username string) (*User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m UserDAOMock) FindByPhone(ctx context.Context, phone string) (*User, error) {
	if m.FindByPhoneFunc != nil {
		return m.FindByPhoneFunc(ctx, phone)
	}
	return nil, nil
}

func (m UserDAOMock) SearchUsers(ctx context.Context, term string, limit int) ([]*User, error) {
	if m.SearchUsersFunc != nil {
		return m.SearchUsersFunc(ctx, term, limit)
	}
	return []*User{}, nil
}

type UserSessionDAOMock struct {
	DAOMock[UserSession]
	FindByRefreshTokenFunc func(ctx context.Context, refreshToken string) (*UserSession, error)
	ListUserSessionsFunc   func(ctx context.Context, userID uint64) ([]*UserSession, error)
	RevokeByUserIDFunc     func(ctx context.Context, userID uint64) error
	RotateRefreshTokenFunc func(ctx context.Context, oldToken string, newSession *UserSession) (*UserSession, error)
}

func (m UserSessionDAOMock) FindByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error) {
	if m.FindByRefreshTokenFunc != nil {
		return m.FindByRefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m UserSessionDAOMock) ListUserSessions(ctx context.Context, userID uint64) ([]*UserSession, error) {
	if m.ListUserSessionsFunc != nil {
		return m.ListUserSessionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m UserSessionDAOMock) RevokeByUserID(ctx context.Context, userID uint64) error {
	if m.RevokeByUserIDFunc != nil {
		return m.RevokeByUserIDFunc(ctx, userID)
	}
	return nil
}

func (m UserSessionDAOMock) RotateRefreshToken(ctx context.Context, oldToken string, newSession *UserSession) (*UserSession, error) {
	if m.RotateRefreshTokenFunc != nil {
		return m.RotateRefreshTokenFunc(ctx, oldToken, newSession)
	}
	return newSession, nil
}

type UserEmailVerificationDAOMock struct {
	DAOMock[UserEmailVerification]
	FindByTokenFunc    func(ctx context.Context, tokenHash string) (*UserEmailVerification, error)
	MarkUsedFunc       func(ctx context.Context, id uint64) error
	DeleteByUserIDFunc func(ctx context.Context, userID uint64) error
}

func (m UserEmailVerificationDAOMock) FindByToken(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
	if m.FindByTokenFunc != nil {
		return m.FindByTokenFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m UserEmailVerificationDAOMock) MarkUsed(ctx context.Context, id uint64) error {
	if m.MarkUsedFunc != nil {
		return m.MarkUsedFunc(ctx, id)
	}
	return nil
}

func (m UserEmailVerificationDAOMock) DeleteByUserID(ctx context.Context, userID uint64) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

type UserPasswordResetDAOMock struct {
	DAOMock[UserPasswordReset]
	FindByTokenHashFunc func(ctx context.Context, tokenHash string) (*UserPasswordReset, error)
	MarkUsedFunc        func(ctx context.Context, id uint64) error
	DeleteByUserIDFunc  func(ctx context.Context, userID uint64) error
}

func (m UserPasswordResetDAOMock) FindByTokenHash(ctx context.Context, tokenHash string) (*UserPasswordReset, error) {
	if m.FindByTokenHashFunc != nil {
		return m.FindByTokenHashFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m UserPasswordResetDAOMock) MarkUsed(ctx context.Context, id uint64) error {
	if m.MarkUsedFunc != nil {
		return m.MarkUsedFunc(ctx, id)
	}
	return nil
}

func (m UserPasswordResetDAOMock) DeleteByUserID(ctx context.Context, userID uint64) error {
	if m.DeleteByUserIDFunc != nil {
		return m.DeleteByUserIDFunc(ctx, userID)
	}
	return nil
}

type PermissionDAOMock struct {
	DAOMock[Permission]
	FindByCodeFunc                      func(ctx context.Context, code string) (*Permission, error)
	UserHasOrganizationPermissionFunc   func(ctx context.Context, userID, organizationID uint64, resource, action string) (bool, error)
	UserIsOrganizationOwnerFunc         func(ctx context.Context, userID, organizationID uint64) (bool, error)
	ListUserOrganizationPermissionsFunc func(ctx context.Context, userID, organizationID uint64) ([]*Permission, error)
}

func (m PermissionDAOMock) FindByCode(ctx context.Context, code string) (*Permission, error) {
	if m.FindByCodeFunc != nil {
		return m.FindByCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m PermissionDAOMock) UserHasOrganizationPermission(ctx context.Context, userID, organizationID uint64, resource, action string) (bool, error) {
	if m.UserHasOrganizationPermissionFunc != nil {
		return m.UserHasOrganizationPermissionFunc(ctx, userID, organizationID, resource, action)
	}
	return false, nil
}

func (m PermissionDAOMock) UserIsOrganizationOwner(ctx context.Context, userID, organizationID uint64) (bool, error) {
	if m.UserIsOrganizationOwnerFunc != nil {
		return m.UserIsOrganizationOwnerFunc(ctx, userID, organizationID)
	}
	return false, nil
}

func (m PermissionDAOMock) ListUserOrganizationPermissions(ctx context.Context, userID, organizationID uint64) ([]*Permission, error) {
	if m.ListUserOrganizationPermissionsFunc != nil {
		return m.ListUserOrganizationPermissionsFunc(ctx, userID, organizationID)
	}
	return nil, nil
}

type MemberDAOMock struct {
	dao.CRUDMock[Member]
	ListOrganizationsByUserFunc   func(ctx context.Context, userID uint64) ([]*reference.Organization, error)
	ListByOrganizationFunc        func(ctx context.Context, organizationID uint64) ([]*Member, error)
	FindByUserAndOrganizationFunc func(ctx context.Context, userID, organizationID uint64) (*Member, error)
	AssignRoleFunc                func(ctx context.Context, memberID, roleID uint64) error
}

func (m MemberDAOMock) ListOrganizationsByUser(ctx context.Context, userID uint64) ([]*reference.Organization, error) {
	if m.ListOrganizationsByUserFunc != nil {
		return m.ListOrganizationsByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m MemberDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Member, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m MemberDAOMock) FindByUserAndOrganization(ctx context.Context, userID, organizationID uint64) (*Member, error) {
	if m.FindByUserAndOrganizationFunc != nil {
		return m.FindByUserAndOrganizationFunc(ctx, userID, organizationID)
	}
	return nil, nil
}

func (m MemberDAOMock) AssignRole(ctx context.Context, memberID, roleID uint64) error {
	if m.AssignRoleFunc != nil {
		return m.AssignRoleFunc(ctx, memberID, roleID)
	}
	return nil
}

type MemberRoleDAOMock struct {
	dao.CRUDMock[MemberRole]
	FindByOrganizationAndCodeFunc func(ctx context.Context, organizationID uint64, code string) (*MemberRole, error)
	ListByOrganizationFunc        func(ctx context.Context, organizationID uint64) ([]*MemberRole, error)
	BindPermissionFunc            func(ctx context.Context, roleID, permissionID uint64) error
	SetPermissionsFunc            func(ctx context.Context, roleID uint64, permissionIDs []uint64) error
}

func (m MemberRoleDAOMock) FindByOrganizationAndCode(ctx context.Context, organizationID uint64, code string) (*MemberRole, error) {
	if m.FindByOrganizationAndCodeFunc != nil {
		return m.FindByOrganizationAndCodeFunc(ctx, organizationID, code)
	}
	return nil, nil
}

func (m MemberRoleDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*MemberRole, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func (m MemberRoleDAOMock) BindPermission(ctx context.Context, roleID, permissionID uint64) error {
	if m.BindPermissionFunc != nil {
		return m.BindPermissionFunc(ctx, roleID, permissionID)
	}
	return nil
}

func (m MemberRoleDAOMock) SetPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	if m.SetPermissionsFunc != nil {
		return m.SetPermissionsFunc(ctx, roleID, permissionIDs)
	}
	return nil
}
