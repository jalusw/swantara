package iam

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const (
	MemberRoleOwner  = "owner"
	MemberRoleMember = "member"
)

type MemberService struct {
	memberDAO     MemberDAO
	memberRoleDAO MemberRoleDAO
	permissionDAO PermissionDAO
}

func NewMemberService(memberDAO MemberDAO, memberRoleDAO MemberRoleDAO, permissionDAO PermissionDAO) MemberService {
	return MemberService{
		memberDAO:     memberDAO,
		memberRoleDAO: memberRoleDAO,
		permissionDAO: permissionDAO,
	}
}

func (s MemberService) CountMembers(ctx context.Context, organizationID uint64) (int, error) {
	members, err := s.memberDAO.ListByOrganization(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

func (s MemberService) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Member, error) {
	return s.memberDAO.ListByOrganization(ctx, organizationID)
}

func (s MemberService) FindByUserAndOrganization(ctx context.Context, userID, organizationID uint64) (*Member, error) {
	return s.memberDAO.FindByUserAndOrganization(ctx, userID, organizationID)
}

func (s MemberService) ListOrganizationsByUser(ctx context.Context, userID uint64) ([]*reference.Organization, error) {
	return s.memberDAO.ListOrganizationsByUser(ctx, userID)
}

func (s MemberService) ListRolesByOrganization(ctx context.Context, organizationID uint64) ([]*MemberRole, error) {
	return s.memberRoleDAO.ListByOrganization(ctx, organizationID)
}

func (s MemberService) ListPermissions(ctx context.Context, q *query.Query) (*query.Page[Permission], error) {
	return s.permissionDAO.List(ctx, q)
}

func (s MemberService) ProvisionOwner(ctx context.Context, organizationID, ownerUserID uint64) error {
	existing, err := s.memberDAO.FindByUserAndOrganization(ctx, ownerUserID, organizationID)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != 0 {
		return nil
	}

	ownerRole, err := s.ensureRole(ctx, organizationID, MemberRoleOwner, "Owner", "Owner with full access to the organization", false)
	if err != nil {
		return err
	}

	if _, err := s.ensureRole(ctx, organizationID, MemberRoleMember, "Member", "Default member with view access to the organization", true); err != nil {
		return err
	}

	member, err := s.memberDAO.Create(ctx, &Member{
		UserID:         ownerUserID,
		OrganizationID: organizationID,
		Position:       helper.Ptr("Owner"),
	})
	if err != nil {
		return err
	}
	return s.memberDAO.AssignRole(ctx, member.ID, ownerRole.ID)
}

func (s MemberService) Invite(ctx context.Context, organizationID, invitedUserID uint64) (*Member, error) {
	existing, err := s.memberDAO.FindByUserAndOrganization(ctx, invitedUserID, organizationID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != 0 {
		return nil, ErrAlreadyMember
	}

	memberRole, err := s.memberRoleDAO.FindByOrganizationAndCode(ctx, organizationID, MemberRoleMember)
	if err != nil {
		return nil, err
	}
	if memberRole == nil {
		return nil, ErrMemberRoleNotFound
	}

	member, err := s.memberDAO.Create(ctx, &Member{UserID: invitedUserID, OrganizationID: organizationID})
	if err != nil {
		return nil, err
	}
	if err := s.memberDAO.AssignRole(ctx, member.ID, memberRole.ID); err != nil {
		return nil, err
	}
	return member, nil
}

func (s MemberService) CreateRole(ctx context.Context, organizationID uint64, role *MemberRole, permissionIDs []uint64) (*MemberRole, error) {
	role.OrganizationID = organizationID

	existing, err := s.memberRoleDAO.FindByOrganizationAndCode(ctx, organizationID, role.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != 0 {
		return nil, ErrMemberRoleExists
	}

	created, err := s.memberRoleDAO.Create(ctx, role)
	if err != nil {
		return nil, err
	}
	if err := s.memberRoleDAO.SetPermissions(ctx, created.ID, permissionIDs); err != nil {
		return nil, err
	}
	return created, nil
}

func (s MemberService) RemoveMember(ctx context.Context, organizationID, memberID uint64) error {
	member, err := s.memberDAO.Find(ctx, memberID)
	if err != nil {
		return err
	}
	if member == nil || member.ID == 0 || member.OrganizationID != organizationID {
		return ErrMemberNotFound
	}
	return s.memberDAO.HardDelete(ctx, memberID)
}

func (s MemberService) ensureRole(ctx context.Context, organizationID uint64, code, name, description string, viewOnly bool) (*MemberRole, error) {
	role, err := s.memberRoleDAO.FindByOrganizationAndCode(ctx, organizationID, code)
	if err != nil {
		return nil, err
	}
	if role != nil && role.ID != 0 {
		return role, nil
	}

	role, err = s.memberRoleDAO.Create(ctx, &MemberRole{
		OrganizationID: organizationID,
		Name:           name,
		Code:           code,
		Description:    helper.Ptr(description),
	})
	if err != nil {
		return nil, err
	}

	if err := s.bindPermissions(ctx, role.ID, viewOnly); err != nil {
		return nil, err
	}
	return role, nil
}

func (s MemberService) bindPermissions(ctx context.Context, roleID uint64, viewOnly bool) error {
	page, err := s.permissionDAO.List(ctx, nil)
	if err != nil {
		return err
	}
	for _, permission := range page.Items {
		if viewOnly && permission.Action != "view" {
			continue
		}
		if err := s.memberRoleDAO.BindPermission(ctx, roleID, permission.ID); err != nil {
			return err
		}
	}
	return nil
}
