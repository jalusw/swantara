package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

func TestAuthN_RequiresActiveSession(t *testing.T) {
	user := iam.UserFixture()
	session := iam.UserSessionFixture(func(s *iam.UserSession) *iam.UserSession {
		s.ID = 42
		s.UserID = user.ID
		return s
	})
	revoked := iam.UserSessionFixture(func(s *iam.UserSession) *iam.UserSession {
		s.ID = 42
		s.UserID = user.ID
		now := time.Now()
		s.RevokedAt = &now
		return s
	})
	foreignSession := iam.UserSessionFixture(func(s *iam.UserSession) *iam.UserSession {
		s.ID = 42
		s.UserID = user.ID + 1
		return s
	})
	expiredSession := iam.UserSessionFixture(func(s *iam.UserSession) *iam.UserSession {
		s.ID = 42
		s.UserID = user.ID
		past := time.Now().Add(-time.Hour)
		s.ExpiresAt = &past
		return s
	})

	cfg := &config.Config{
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
	}
	tokenSvc := iam.NewTokenService(cfg)
	accessToken, _, err := tokenSvc.IssueUserAccessToken(user, 42)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		findUserFunc    func(ctx context.Context, id uint64) (*iam.User, error)
		findSessionFunc func(ctx context.Context, id uint64) (*iam.UserSession, error)
		wantStatus      int
	}{
		{
			name: "active session",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return session, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "revoked session",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return revoked, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing session",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return nil, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "session of another user",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return foreignSession, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired session",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return expiredSession, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "session lookup error",
			findUserFunc: func(ctx context.Context, id uint64) (*iam.User, error) {
				return user, nil
			},
			findSessionFunc: func(ctx context.Context, id uint64) (*iam.UserSession, error) {
				return nil, context.Canceled
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthenticationMiddleware{
				tokenSvc:       tokenSvc,
				userDAO:        iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: tt.findUserFunc}},
				userSessionDAO: iam.UserSessionDAOMock{DAOMock: iam.DAOMock[iam.UserSession]{FindFunc: tt.findSessionFunc}},
			}

			app := fiber.New()
			app.Get("/", middleware.AuthN, func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func authzTestMiddleware(permissionFn func(ctx context.Context, userID, organizationID uint64, resource, action string) (bool, error), ownerFn func(ctx context.Context, userID, organizationID uint64) (bool, error)) AuthorizationMiddleware {
	return AuthorizationMiddleware{
		permissionDAO: iam.PermissionDAOMock{
			UserHasOrganizationPermissionFunc: permissionFn,
			UserIsOrganizationOwnerFunc:       ownerFn,
		},
	}
}

func TestGuard(t *testing.T) {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(LocalUserID, uint64(7))
		return c.Next()
	})

	middleware := authzTestMiddleware(
		func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return true, nil },
		func(_ context.Context, _, _ uint64) (bool, error) { return false, nil },
	)
	app.Get("/guard/:organization_id", middleware.Guard("sales.order", "read"), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	if resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/guard/5", nil)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	app2 := fiber.New()
	app2.Use(func(c fiber.Ctx) error {
		c.Locals(LocalUserID, uint64(7))
		return c.Next()
	})
	denied := authzTestMiddleware(
		func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return false, nil },
		func(_ context.Context, _, _ uint64) (bool, error) { return false, nil },
	)
	app2.Get("/guard/:organization_id", denied.Guard("sales.order", "read"), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	if resp, err := app2.Test(httptest.NewRequest(http.MethodGet, "/guard/5", nil)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusForbidden {
		t.Errorf("denied status = %d, want 403", resp.StatusCode)
	}
}

func TestGuard_OwnerShortCircuit(t *testing.T) {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(LocalUserID, uint64(7))
		return c.Next()
	})
	owner := authzTestMiddleware(
		func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return false, nil },
		func(_ context.Context, _, _ uint64) (bool, error) { return true, nil },
	)
	app.Get("/guard/:organization_id", owner.Guard("sales.order", "read"), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	if resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/guard/5", nil)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Errorf("owner status = %d, want 200", resp.StatusCode)
	}
}

func TestAuthzConstructors(t *testing.T) {
	authn := NewAuthenticationMiddleware(iam.TokenService{}, nil, nil)
	_ = authn
	authz := NewAuthorizationMiddleware(iam.PermissionDAOMock{})
	if authz.permissionDAO == nil {
		t.Error("NewAuthorizationMiddleware() did not store permission dao")
	}
}

func TestAuthN_MissingHeader(t *testing.T) {
	middleware := AuthenticationMiddleware{}
	app := fiber.New()
	app.Get("/", middleware.AuthN)
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthN_InvalidToken(t *testing.T) {
	tokenSvc := iam.NewTokenService(&config.Config{
		JWTAccessTokenSecret:           "secret",
		JWTAccessTokenExpires:          "30m",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh",
		JWTRefreshTokenExpires:         "1h",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})
	middleware := AuthenticationMiddleware{tokenSvc: tokenSvc}
	app := fiber.New()
	app.Get("/", middleware.AuthN)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}
