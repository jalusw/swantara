package httpx

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type AuthenticationMiddleware struct {
	tokenSvc       iam.TokenService
	userDAO        iam.UserDAO
	userSessionDAO iam.UserSessionDAO
}

func NewAuthenticationMiddleware(tokenSvc iam.TokenService, userDAO iam.UserDAO, userSessionDAO iam.UserSessionDAO) AuthenticationMiddleware {
	return AuthenticationMiddleware{tokenSvc: tokenSvc, userDAO: userDAO, userSessionDAO: userSessionDAO}
}

func (m AuthenticationMiddleware) AuthN(c fiber.Ctx) error {
	accessToken := c.Get("Authorization")

	if accessToken == "" {
		RequestLog(c).Warn("authn_rejected", "reason", "missing_authorization_header")
		return CreateUnauthorizedErrorResponse(c, "Unauthorized.", ErrMissingAuthorizationHeader)
	}

	tokenString := strings.TrimPrefix(accessToken, "Bearer ")

	claims, err := m.tokenSvc.VerifyUserAccessToken(tokenString)

	if err != nil {
		RequestLog(c).Warn("authn_rejected", "reason", "token_invalid", "error", err)
		return CreateUnauthorizedErrorResponse(c, "Invalid token.", err)
	}

	user, err := m.userDAO.Find(c, claims.UserID)
	if err != nil {
		RequestLog(c).Error("user find failed during authentication", "user_id", claims.UserID, "error", err)
		return CreateInternalServerErrorResponse(c, "", err)
	}

	if user == nil || user.ID == 0 || !user.Active {
		RequestLog(c).Warn("authn_rejected", "reason", "user_inactive", "user_id", claims.UserID)
		return CreateUnauthorizedErrorResponse(c, "Unauthorized.", iam.ErrUserNotActive)
	}

	session, err := m.userSessionDAO.Find(c, claims.SessionID)
	if err != nil {
		RequestLog(c).Error("session find failed during authentication", "session_id", claims.SessionID, "error", err)
		return CreateInternalServerErrorResponse(c, "", err)
	}

	if session == nil || session.UserID != user.ID || session.RevokedAt != nil {
		RequestLog(c).Warn("authn_rejected", "reason", "session_invalid", "user_id", user.ID, "session_id", claims.SessionID)
		return CreateUnauthorizedErrorResponse(c, "Unauthorized.", iam.ErrSessionRevoked)
	}

	if session.ExpiresAt != nil && time.Now().After(*session.ExpiresAt) {
		RequestLog(c).Warn("authn_rejected", "reason", "session_expired", "user_id", user.ID, "session_id", claims.SessionID)
		return CreateUnauthorizedErrorResponse(c, "Unauthorized.", iam.ErrSessionExpired)
	}

	c.Locals(LocalUserID, user.ID)

	return c.Next()
}

type AuthorizationMiddleware struct {
	permissionDAO iam.PermissionDAO
}

func NewAuthorizationMiddleware(permissionDAO iam.PermissionDAO) AuthorizationMiddleware {
	return AuthorizationMiddleware{permissionDAO: permissionDAO}
}

func (m AuthorizationMiddleware) Guard(resource, action string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := CallerID(c)
		if !ok {
			RequestLog(c).Warn("authn_rejected", "reason", "missing_user_in_context")
			return CreateUnauthorizedErrorResponse(c, "Unauthorized.", ErrMissingUserInContext)
		}

		organizationID, ok := ResolveOrganizationID(c)
		if !ok {
			RequestLog(c).Warn("tenant_rejected", "reason", "missing_organization_in_context")
			return CreateUnauthorizedErrorResponse(c, "Unauthorized.", ErrMissingOrganizationInContext)
		}
		c.Locals(LocalOrganizationID, organizationID)

		if m.isOrganizationOwner(c, userID, organizationID) {
			return c.Next()
		}

		allowed, err := m.permissionDAO.UserHasOrganizationPermission(c, userID, organizationID, resource, action)
		if err != nil {
			RequestLog(c).Error("organization permission check failed", "user_id", userID, "organization_id", organizationID, "resource", resource, "action", action, "error", err)
			return CreateInternalServerErrorResponse(c, "", err)
		}

		if !allowed {
			RequestLog(c).Warn("authz_denied", "resource", resource, "action", action,
				"user_id", userID, "organization_id", organizationID)
			return CreateForbiddenErrorResponse(c, "Forbidden.", iam.ErrForbidden)
		}

		return c.Next()
	}
}

func (m AuthorizationMiddleware) isOrganizationOwner(c fiber.Ctx, userID, organizationID uint64) bool {
	owner, err := m.permissionDAO.UserIsOrganizationOwner(c, userID, organizationID)
	if err != nil {
		RequestLog(c).Warn("authz_owner_check_failed", "error", err)
		return false
	}
	return owner
}
