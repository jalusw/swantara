package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

func testPasswordService() iam.PasswordService {
	return iam.NewPasswordService(&config.Config{BcryptCost: bcrypt.MinCost, Pepper: "test-pepper"})
}

func testTokenService() iam.TokenService {
	return iam.NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "access-secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})
}

func authTestService(
	users iam.UserDAOMock,
	sessions iam.UserSessionDAOMock,
	emailVerifications iam.UserEmailVerificationDAOMock,
	passwordResets iam.UserPasswordResetDAOMock,
	queueClient queue.TaskEnqueuer,
) iam.AuthService {
	passwordSvc := testPasswordService()
	return iam.NewAuthService(
		users,
		iam.NewUserService(users, passwordSvc),
		sessions,
		passwordSvc,
		testTokenService(),
		emailVerifications,
		passwordResets,
		queueClient,
		time.Hour*24,
		time.Minute*30,
	)
}

func sampleUser() *iam.User {
	return &iam.User{
		Base:      model.Base{ID: 1},
		Username:  "jane.doe",
		FirstName: "Jane",
		LastName:  helper.Ptr("Doe"),
		Email:     "jane@example.com",
		Phone:     helper.Ptr("+6281234567890"),
		Active:    true,
	}
}

func sampleUserWithPassword() *iam.User {
	user := sampleUser()
	hash, _ := testPasswordService().Hash("password123")
	user.Password = hash
	return user
}

func sampleMember() *iam.Member {
	return &iam.Member{
		Base:           model.Base{ID: 1},
		UserID:         5,
		OrganizationID: 10,
		Position:       helper.Ptr("Owner"),
	}
}

func sampleMemberRole() *iam.MemberRole {
	return &iam.MemberRole{
		Base:           model.Base{ID: 1},
		OrganizationID: 10,
		Name:           "Member",
		Code:           "member",
	}
}

func samplePermission() *iam.Permission {
	return &iam.Permission{
		Base:     model.Base{ID: 1},
		Name:     "View Members",
		Code:     "member:view",
		Action:   "view",
		Resource: "member",
	}
}

func sampleSession() *iam.UserSession {
	return &iam.UserSession{
		Base:       model.Base{ID: 1},
		UserID:     5,
		DeviceName: helper.Ptr("Chrome"),
		OS:         helper.Ptr("Linux"),
		Browser:    helper.Ptr("Chrome"),
		IPAddress:  helper.Ptr("127.0.0.1"),
	}
}

func sampleOrganization() *reference.Organization {
	return &reference.Organization{Base: model.Base{ID: 10}, Name: "Acme"}
}

func validRefreshToken() string {
	token, _, _ := testTokenService().IssueUserRefreshToken(sampleUser())
	return token
}

func authHandlerTest(t *testing.T, authSvc iam.AuthService, userDAO iam.UserDAO, sessions iam.UserSessionDAO) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewAuthHandler(authSvc)
	h.Register(app, passthroughGuards())
	return app
}

func authHandlerTestNoCaller(t *testing.T, authSvc iam.AuthService, userDAO iam.UserDAO, sessions iam.UserSessionDAO) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewAuthHandler(authSvc)
	h.Register(app, passthroughGuards())
	return app
}

func memberHandlerTest(t *testing.T, members iam.MemberDAOMock, roles iam.MemberRoleDAOMock, permissions iam.PermissionDAOMock, withTenant bool) *fiber.App {
	t.Helper()
	app := fiber.New()
	if withTenant {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(10))
			return c.Next()
		})
	}
	memberSvc := iam.NewMemberService(members, roles, permissions)
	h := NewMemberHandler(memberSvc)
	h.Register(app, passthroughGuards())
	return app
}

func userHandlerTest(t *testing.T, users iam.UserDAOMock, permissions iam.PermissionDAOMock, members iam.MemberDAOMock, actorID uint64) *fiber.App {
	t.Helper()
	return userHandlerTestApp(t, users, permissions, members, actorID, avatarHandlerTestService(t, users))
}

func userHandlerTestNoCaller(t *testing.T, users iam.UserDAOMock, permissions iam.PermissionDAOMock, members iam.MemberDAOMock) *fiber.App {
	t.Helper()
	return userHandlerTestApp(t, users, permissions, members, 0, avatarHandlerTestService(t, users))
}

func userHandlerTestApp(
	t *testing.T,
	users iam.UserDAOMock,
	permissions iam.PermissionDAOMock,
	members iam.MemberDAOMock,
	actorID uint64,
	avatarSvc iam.AvatarService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	if actorID != 0 {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(model.ActorKey, actorID)
			return c.Next()
		})
	}
	userSvc := iam.NewUserService(users, testPasswordService())
	authzSvc := iam.NewAuthzService(permissions)
	memberSvc := iam.NewMemberService(members, iam.MemberRoleDAOMock{}, iam.PermissionDAOMock{})
	h := NewUserHandler(userSvc, authzSvc, memberSvc, avatarSvc)
	h.Register(app, passthroughGuards())
	return app
}

func avatarHandlerTestService(t *testing.T, users iam.UserDAO) iam.AvatarService {
	t.Helper()
	return iam.NewAvatarService(users, storage.NewLocalStore(t.TempDir()))
}
