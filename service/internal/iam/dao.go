package iam

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserDAO interface {
	dao.CRUD[User]
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	SearchUsers(ctx context.Context, term string, limit int) ([]*User, error)
}

type userDAO struct {
	dao.Base[User]
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return userDAO{Base: dao.NewBase[User](db), db: db}
}

func (s userDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
	return s.Search(ctx, "email", email)
}

func (s userDAO) FindByUsername(ctx context.Context, username string) (*User, error) {
	return s.Search(ctx, "username", username)
}

func (s userDAO) FindByPhone(ctx context.Context, phone string) (*User, error) {
	return s.Search(ctx, "phone", phone)
}

func (s userDAO) SearchUsers(ctx context.Context, term string, limit int) ([]*User, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	pattern := "%" + term + "%"
	var users []*User
	err := s.db.WithContext(ctx).
		Where("private = ?", false).
		Where(
			s.db.WithContext(ctx).
				Where("username ILIKE ?", pattern).
				Or("first_name ILIKE ?", pattern).
				Or("last_name ILIKE ?", pattern).
				Or("CONCAT(first_name, ' ', COALESCE(last_name, '')) ILIKE ?", pattern),
		).
		Order("username ASC").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

type UserSessionDAO interface {
	dao.CRUD[UserSession]
	FindByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error)
	ListUserSessions(ctx context.Context, userID uint64) ([]*UserSession, error)
	RevokeByUserID(ctx context.Context, userID uint64) error
	RotateRefreshToken(ctx context.Context, oldToken string, newSession *UserSession) (*UserSession, error)
}

type userSessionDAO struct {
	dao.Base[UserSession]
	db *gorm.DB
}

func NewUserSessionDAO(db *gorm.DB) UserSessionDAO {
	return userSessionDAO{Base: dao.NewBase[UserSession](db), db: db}
}

func (s userSessionDAO) FindByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error) {
	var session *UserSession
	if err := s.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).Find(&session).Error; err != nil {
		return nil, err
	}
	if session == nil || session.ID == 0 {
		return nil, nil
	}
	return session, nil
}

func (s userSessionDAO) ListUserSessions(ctx context.Context, userID uint64) ([]*UserSession, error) {
	var sessions []*UserSession
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s userSessionDAO) RevokeByUserID(ctx context.Context, userID uint64) error {
	return s.db.WithContext(ctx).Model(&UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error
}

const refreshReuseGracePeriod = 60 * time.Second

func (s userSessionDAO) RotateRefreshToken(ctx context.Context, oldToken string, newSession *UserSession) (*UserSession, error) {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var oldSession UserSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refresh_token = ?", oldToken).
			First(&oldSession).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTokenNotFound
			}
			return err
		}

		if oldSession.RevokedAt != nil {
			if time.Since(*oldSession.RevokedAt) > refreshReuseGracePeriod {
				return ErrTokenReused
			}
			return tx.Create(newSession).Error
		}

		now := time.Now()
		oldSession.RevokedAt = &now
		if err := tx.Save(&oldSession).Error; err != nil {
			return err
		}

		return tx.Create(newSession).Error
	}); err != nil {
		return nil, err
	}

	return newSession, nil
}

type UserEmailVerificationDAO interface {
	dao.CRUD[UserEmailVerification]
	FindByToken(ctx context.Context, tokenHash string) (*UserEmailVerification, error)
	MarkUsed(ctx context.Context, id uint64) error
	DeleteByUserID(ctx context.Context, userID uint64) error
}

type userEmailVerificationDAO struct {
	dao.Base[UserEmailVerification]
	db *gorm.DB
}

func NewUserEmailVerificationDAO(db *gorm.DB) UserEmailVerificationDAO {
	return userEmailVerificationDAO{Base: dao.NewBase[UserEmailVerification](db), db: db}
}

func (s userEmailVerificationDAO) FindByToken(ctx context.Context, tokenHash string) (*UserEmailVerification, error) {
	var entity *UserEmailVerification
	if err := s.db.WithContext(ctx).Where("verification_token = ?", tokenHash).Find(&entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (s userEmailVerificationDAO) MarkUsed(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Model(&UserEmailVerification{}).
		Where("id = ?", id).
		Update("used_at", time.Now()).Error
}

func (s userEmailVerificationDAO) DeleteByUserID(ctx context.Context, userID uint64) error {
	return s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserEmailVerification{}).Error
}

type UserPasswordResetDAO interface {
	dao.CRUD[UserPasswordReset]
	FindByTokenHash(ctx context.Context, tokenHash string) (*UserPasswordReset, error)
	MarkUsed(ctx context.Context, id uint64) error
	DeleteByUserID(ctx context.Context, userID uint64) error
}

type userPasswordResetDAO struct {
	dao.Base[UserPasswordReset]
	db *gorm.DB
}

func NewUserPasswordResetDAO(db *gorm.DB) UserPasswordResetDAO {
	return userPasswordResetDAO{Base: dao.NewBase[UserPasswordReset](db), db: db}
}

func (s userPasswordResetDAO) FindByTokenHash(ctx context.Context, tokenHash string) (*UserPasswordReset, error) {
	var entity *UserPasswordReset
	if err := s.db.WithContext(ctx).Where("reset_token_hash = ?", tokenHash).Find(&entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (s userPasswordResetDAO) MarkUsed(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Model(&UserPasswordReset{}).
		Where("id = ?", id).
		Update("used_at", time.Now()).Error
}

func (s userPasswordResetDAO) DeleteByUserID(ctx context.Context, userID uint64) error {
	return s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserPasswordReset{}).Error
}

type PermissionDAO interface {
	dao.CRUD[Permission]
	FindByCode(ctx context.Context, code string) (*Permission, error)
	UserHasOrganizationPermission(ctx context.Context, userID, organizationID uint64, resource, action string) (bool, error)
	UserIsOrganizationOwner(ctx context.Context, userID, organizationID uint64) (bool, error)
	ListUserOrganizationPermissions(ctx context.Context, userID, organizationID uint64) ([]*Permission, error)
}

type permissionDAO struct {
	dao.Base[Permission]
	db *gorm.DB
}

func NewPermissionDAO(db *gorm.DB) PermissionDAO {
	return permissionDAO{Base: dao.NewBase[Permission](db), db: db}
}

func (s permissionDAO) FindByCode(ctx context.Context, code string) (*Permission, error) {
	var entity *Permission
	if err := s.db.WithContext(ctx).Where("code = ?", code).Find(&entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (s permissionDAO) UserHasOrganizationPermission(ctx context.Context, userID, organizationID uint64, resource, action string) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Table("members").
		Joins("JOIN member_role_assignments ON member_role_assignments.member_id = members.id").
		Joins("JOIN member_roles ON member_roles.id = member_role_assignments.role_id").
		Joins("JOIN member_role_permissions ON member_role_permissions.member_role_id = member_roles.id").
		Joins("JOIN permissions ON permissions.id = member_role_permissions.permission_id").
		Where("members.user_id = ?", userID).
		Where("members.organization_id = ?", organizationID).
		Where("member_roles.organization_id = ?", organizationID).
		Where("permissions.resource = ?", resource).
		Where("permissions.action = ?", action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s permissionDAO) UserIsOrganizationOwner(ctx context.Context, userID, organizationID uint64) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Table("members").
		Joins("JOIN member_role_assignments ON member_role_assignments.member_id = members.id").
		Joins("JOIN member_roles ON member_roles.id = member_role_assignments.role_id").
		Where("members.user_id = ?", userID).
		Where("members.organization_id = ?", organizationID).
		Where("member_roles.code = ?", MemberRoleOwner).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s permissionDAO) ListUserOrganizationPermissions(ctx context.Context, userID, organizationID uint64) ([]*Permission, error) {
	var permissions []*Permission

	err := s.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN member_role_permissions ON member_role_permissions.permission_id = permissions.id").
		Joins("JOIN member_roles ON member_roles.id = member_role_permissions.member_role_id").
		Joins("JOIN member_role_assignments ON member_role_assignments.role_id = member_roles.id").
		Joins("JOIN members ON members.id = member_role_assignments.member_id").
		Where("members.user_id = ?", userID).
		Where("members.organization_id = ?", organizationID).
		Where("member_roles.organization_id = ?", organizationID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

type MemberDAO interface {
	dao.CRUD[Member]
	ListOrganizationsByUser(ctx context.Context, userID uint64) ([]*reference.Organization, error)
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*Member, error)
	FindByUserAndOrganization(ctx context.Context, userID, organizationID uint64) (*Member, error)
	AssignRole(ctx context.Context, memberID, roleID uint64) error
}

type memberDAO struct {
	dao.Base[Member]
	db *gorm.DB
}

func NewMemberDAO(db *gorm.DB) MemberDAO {
	return memberDAO{Base: dao.NewBase[Member](db), db: db}
}

func (s memberDAO) ListOrganizationsByUser(ctx context.Context, userID uint64) ([]*reference.Organization, error) {
	var members []*Member
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Organization").
		Find(&members).Error; err != nil {
		return nil, err
	}

	orgs := make([]*reference.Organization, len(members))
	for i, member := range members {
		orgs[i] = member.Organization
	}
	return orgs, nil
}

func (s memberDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Member, error) {
	var members []*Member
	if err := s.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Preload("User").
		Preload("Roles").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (s memberDAO) FindByUserAndOrganization(ctx context.Context, userID, organizationID uint64) (*Member, error) {
	var member *Member
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return member, nil
}

func (s memberDAO) AssignRole(ctx context.Context, memberID, roleID uint64) error {
	return s.db.WithContext(ctx).
		Table("member_role_assignments").
		Create(map[string]any{"member_id": memberID, "role_id": roleID}).Error
}

type MemberRoleDAO interface {
	dao.CRUD[MemberRole]
	FindByOrganizationAndCode(ctx context.Context, organizationID uint64, code string) (*MemberRole, error)
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*MemberRole, error)
	BindPermission(ctx context.Context, roleID, permissionID uint64) error
	SetPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
}

type memberRoleDAO struct {
	dao.Base[MemberRole]
	db *gorm.DB
}

func NewMemberRoleDAO(db *gorm.DB) MemberRoleDAO {
	return memberRoleDAO{Base: dao.NewBase[MemberRole](db), db: db}
}

func (s memberRoleDAO) FindByOrganizationAndCode(ctx context.Context, organizationID uint64, code string) (*MemberRole, error) {
	var role *MemberRole
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND code = ?", organizationID, code).
		First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func (s memberRoleDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*MemberRole, error) {
	var roles []*MemberRole
	if err := s.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Preload("Permissions").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s memberRoleDAO) BindPermission(ctx context.Context, roleID, permissionID uint64) error {
	return s.db.WithContext(ctx).
		Table("member_role_permissions").
		Create(map[string]any{"member_role_id": roleID, "permission_id": permissionID}).Error
}

func (s memberRoleDAO) SetPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	if err := s.db.WithContext(ctx).Unscoped().Where("member_role_id = ?", roleID).Delete(&MemberRolePermission{}).Error; err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if err := s.BindPermission(ctx, roleID, permissionID); err != nil {
			return err
		}
	}
	return nil
}
