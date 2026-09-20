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
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

func extraTokenService() iam.TokenService {
	return iam.NewTokenService(&config.Config{
		ApplicationURL:                 "http://localhost:8080",
		ClientWebURL:                   "http://localhost:3000",
		JWTAccessTokenSecret:           "access-secret",
		JWTAccessTokenExpiresDuration:  30 * time.Minute,
		JWTAccessTokenSigningMethod:    jwt.SigningMethodHS256,
		JWTRefreshTokenSecret:          "refresh-secret",
		JWTRefreshTokenExpiresDuration: time.Hour,
		JWTRefreshTokenSigningMethod:   jwt.SigningMethodHS256,
	})
}

func TestAuthN_UserBranches(t *testing.T) {
	tokenSvc := extraTokenService()
	user := iam.UserFixture()
	accessToken, _, err := tokenSvc.IssueUserAccessToken(user, 42)
	if err != nil {
		t.Fatal(err)
	}

	run := func(userDAO iam.UserDAO, sessionDAO iam.UserSessionDAO) int {
		middleware := AuthenticationMiddleware{tokenSvc: tokenSvc, userDAO: userDAO, userSessionDAO: sessionDAO}
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
		return resp.StatusCode
	}

	t.Run("user lookup error", func(t *testing.T) {
		dao := iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{
			FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return nil, context.Canceled },
		}}
		if got := run(dao, iam.UserSessionDAOMock{}); got != http.StatusInternalServerError {
			t.Errorf("status = %d", got)
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		inactive := *user
		inactive.Active = false
		dao := iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{
			FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return &inactive, nil },
		}}
		if got := run(dao, iam.UserSessionDAOMock{}); got != http.StatusUnauthorized {
			t.Errorf("status = %d", got)
		}
	})
}

func TestGuard_Branches(t *testing.T) {
	t.Run("rejects missing user and org", func(t *testing.T) {
		middleware := authzTestMiddleware(
			func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return true, nil },
			func(_ context.Context, _, _ uint64) (bool, error) { return false, nil },
		)
		app := fiber.New()
		app.Get("/", middleware.Guard("item", "view"), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		appWithUser := fiber.New()
		appWithUser.Use(func(c fiber.Ctx) error {
			c.Locals(LocalUserID, uint64(7))
			return c.Next()
		})
		appWithUser.Get("/", middleware.Guard("item", "view"), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})
		resp, err = appWithUser.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)
	})

	t.Run("reports permission error", func(t *testing.T) {
		middleware := authzTestMiddleware(
			func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return false, context.Canceled },
			func(_ context.Context, _, _ uint64) (bool, error) { return false, nil },
		)
		app := fiber.New()
		app.Use(func(c fiber.Ctx) error {
			c.Locals(LocalUserID, uint64(7))
			return c.Next()
		})
		app.Get("/", middleware.Guard("item", "view"), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(OrganizationHeader, "10")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("owner check error falls through", func(t *testing.T) {
		middleware := authzTestMiddleware(
			func(_ context.Context, _, _ uint64, _, _ string) (bool, error) { return true, nil },
			func(_ context.Context, _, _ uint64) (bool, error) { return false, context.Canceled },
		)
		app := fiber.New()
		app.Use(func(c fiber.Ctx) error {
			c.Locals(LocalUserID, uint64(7))
			return c.Next()
		})
		app.Get("/", middleware.Guard("item", "view"), func(c fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(OrganizationHeader, "10")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}
