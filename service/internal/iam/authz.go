package iam

import (
	"context"
)

type AuthzService struct {
	permissionDAO PermissionDAO
}

func NewAuthzService(
	permissionDAO PermissionDAO,
) AuthzService {
	return AuthzService{
		permissionDAO: permissionDAO,
	}
}

func (s AuthzService) EffectiveOrganizationPermissions(
	ctx context.Context,
	userID, organizationID uint64,
) ([]*Permission, error) {
	owner, err := s.permissionDAO.UserIsOrganizationOwner(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}

	if owner {
		page, err := s.permissionDAO.List(ctx, nil)
		if err != nil {
			return nil, err
		}
		return page.Items, nil
	}

	return s.permissionDAO.ListUserOrganizationPermissions(ctx, userID, organizationID)
}
