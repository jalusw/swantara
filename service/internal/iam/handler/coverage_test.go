package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/queue"
)

func TestAuthHandler_CheckEmail(t *testing.T) {
	newApp := func(users iam.UserDAOMock) *fiber.App {
		t.Helper()
		app := fiber.New()
		svc := authTestService(users, iam.UserSessionDAOMock{}, iam.UserEmailVerificationDAOMock{}, iam.UserPasswordResetDAOMock{}, queue.TaskEnqueuerMock{})
		h := NewAuthHandler(svc)
		h.Register(app, passthroughGuards())
		return app
	}

	t.Run("reports availability", func(t *testing.T) {
		app := newApp(iam.UserDAOMock{})
		resp, err := doIamRequest(app, http.MethodPost, "/auth/email/check", `{"email":"new@example.com"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("reports taken", func(t *testing.T) {
		app := newApp(iam.UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
				return sampleUser(), nil
			},
		})
		resp, err := doIamRequest(app, http.MethodPost, "/auth/email/check", `{"email":"taken@example.com"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid body and reports error", func(t *testing.T) {
		app := newApp(iam.UserDAOMock{})
		resp, err := doIamRequest(app, http.MethodPost, "/auth/email/check", `{"email":"bad"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		failing := newApp(iam.UserDAOMock{
			FindByEmailFunc: func(_ context.Context, _ string) (*iam.User, error) {
				return nil, errors.New("db down")
			},
		})
		resp, err = doIamRequest(failing, http.MethodPost, "/auth/email/check", `{"email":"a@b.c"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestWriteAccessError(t *testing.T) {
	app := fiber.New()
	app.Get("/probe/:kind", func(c fiber.Ctx) error {
		switch c.Params("kind") {
		case "forbidden":
			return writeAccessError(c, iam.ErrForbidden)
		case "missing":
			return writeAccessError(c, httpx.ErrMissingUserInContext)
		default:
			return writeAccessError(c, errors.New("boom"))
		}
	})

	resp, err := doIamRequest(app, http.MethodGet, "/probe/forbidden", "")
	if err != nil {
		t.Fatal(err)
	}
	helper.AssertStatus(t, resp.StatusCode, http.StatusForbidden)

	resp, err = doIamRequest(app, http.MethodGet, "/probe/missing", "")
	if err != nil {
		t.Fatal(err)
	}
	helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

	resp, err = doIamRequest(app, http.MethodGet, "/probe/other", "")
	if err != nil {
		t.Fatal(err)
	}
	helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
}

func doIamRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
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
