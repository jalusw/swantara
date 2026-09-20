//go:build e2e

package e2e

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

type authTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type registeredUser struct {
	email    string
	password string
	tokens   authTokens
	ID       uint64
}

func registerAndLogin(t *testing.T) registeredUser {
	t.Helper()

	e := newExpect(t)

	email := gofakeit.Email()
	password := "test-password"

	registerUser := e.POST("/api/v1/auth/register").
		WithJSON(map[string]any{
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      email,
			"password":   password,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("user").
		Object()

	loginData := e.POST("/api/v1/auth/login").
		WithJSON(map[string]any{
			"email":    email,
			"password": password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object()

	user := registeredUser{
		email:    email,
		password: password,
		ID:       uint64(registerUser.Value("id").Number().Raw()),
		tokens: authTokens{
			AccessToken:  loginData.Value("access_token").String().Raw(),
			RefreshToken: loginData.Value("refresh_token").String().Raw(),
		},
	}

	t.Logf("e2e: created user id=%d email=%s password=%s", user.ID, user.email, user.password)
	t.Logf("e2e: user tokens access_token=%s refresh_token=%s", user.tokens.AccessToken, user.tokens.RefreshToken)

	return user
}

func TestRegisterLoginE2E(t *testing.T) {
	user := registerAndLogin(t)

	if user.tokens.AccessToken == "" || user.tokens.RefreshToken == "" {
		t.Error("expected access and refresh tokens")
	}

	newExpect(t).GET("/api/v1/users").
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusOK)
}

func TestRegisterDuplicateEmailE2E(t *testing.T) {
	user := registerAndLogin(t)

	newExpect(t).POST("/api/v1/auth/register").
		WithJSON(map[string]any{
			"first_name": gofakeit.FirstName(),
			"email":      user.email,
			"password":   "test-password",
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestLoginWrongPasswordE2E(t *testing.T) {
	user := registerAndLogin(t)

	newExpect(t).POST("/api/v1/auth/login").
		WithJSON(map[string]any{
			"email":    user.email,
			"password": "wrong-password",
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRefreshAfterLogoutRejectedE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	e.POST("/api/v1/auth/logout").
		WithJSON(map[string]any{
			"refresh_token": user.tokens.RefreshToken,
		}).
		Expect().
		Status(http.StatusNoContent)

	e.POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{
			"refresh_token": user.tokens.RefreshToken,
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestPublicUserReadE2E(t *testing.T) {
	newExpect(t).GET("/api/v1/users/1").
		Expect().
		Status(http.StatusOK)
}

func TestRegisterValidationErrorsE2E(t *testing.T) {
	e := newExpect(t)

	e.POST("/api/v1/auth/register").
		WithJSON(map[string]any{
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      gofakeit.Email(),
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/v1/auth/register").
		WithJSON(map[string]any{
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      "not-an-email",
			"password":   "test-password",
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/v1/auth/register").
		WithJSON(map[string]any{
			"first_name": gofakeit.FirstName(),
			"last_name":  gofakeit.LastName(),
			"email":      gofakeit.Email(),
			"password":   "short",
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestLoginValidationErrorsE2E(t *testing.T) {
	e := newExpect(t)

	e.POST("/api/v1/auth/login").
		WithJSON(map[string]any{"email": gofakeit.Email()}).
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/v1/auth/login").
		WithJSON(map[string]any{"password": "test-password"}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestLoginUnknownEmailE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/login").
		WithJSON(map[string]any{
			"email":    gofakeit.Email(),
			"password": "test-password",
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRefreshRotatesTokensE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	rotated := e.POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{"refresh_token": user.tokens.RefreshToken}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object()

	newAccessToken := rotated.Value("access_token").String().Raw()
	newRefreshToken := rotated.Value("refresh_token").String().Raw()

	if newAccessToken == "" || newRefreshToken == "" {
		t.Error("expected fresh access and refresh tokens")
	}
	if newAccessToken == user.tokens.AccessToken {
		t.Error("expected access token to be rotated")
	}
	if newRefreshToken == user.tokens.RefreshToken {
		t.Error("expected refresh token to be rotated")
	}

	e.POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{"refresh_token": user.tokens.RefreshToken}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRefreshInvalidTokenE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{"refresh_token": "not-a-real-token"}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRefreshValidationErrorsE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestLogoutIdempotentE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	e.POST("/api/v1/auth/logout").
		WithJSON(map[string]any{"refresh_token": user.tokens.RefreshToken}).
		Expect().
		Status(http.StatusNoContent)

	e.POST("/api/v1/auth/logout").
		WithJSON(map[string]any{"refresh_token": user.tokens.RefreshToken}).
		Expect().
		Status(http.StatusNoContent)

	e.POST("/api/v1/auth/logout").
		WithJSON(map[string]any{"refresh_token": "not-a-real-token"}).
		Expect().
		Status(http.StatusNoContent)
}

func TestEmailVerificationRequestNoEnumerationE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	e.POST("/api/v1/auth/email-verification/request").
		WithJSON(map[string]any{"email": user.email}).
		Expect().
		Status(http.StatusAccepted)

	e.POST("/api/v1/auth/email-verification/request").
		WithJSON(map[string]any{"email": gofakeit.Email()}).
		Expect().
		Status(http.StatusAccepted)
}

func TestEmailVerificationRequestValidationErrorsE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/email-verification/request").
		WithJSON(map[string]any{"email": "not-an-email"}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestVerifyEmailInvalidTokenE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/email-verification/verify").
		WithJSON(map[string]any{"token": "not-a-real-token"}).
		Expect().
		Status(http.StatusBadRequest)
}

func TestVerifyEmailValidationErrorsE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/email-verification/verify").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestPasswordResetRequestNoEnumerationE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	e.POST("/api/v1/auth/password-reset/request").
		WithJSON(map[string]any{"email": user.email}).
		Expect().
		Status(http.StatusAccepted)

	e.POST("/api/v1/auth/password-reset/request").
		WithJSON(map[string]any{"email": gofakeit.Email()}).
		Expect().
		Status(http.StatusAccepted)
}

func TestResetPasswordInvalidTokenE2E(t *testing.T) {
	newExpect(t).POST("/api/v1/auth/password-reset").
		WithJSON(map[string]any{
			"token":    "not-a-real-token",
			"password": "new-password",
		}).
		Expect().
		Status(http.StatusBadRequest)
}

func TestResetPasswordValidationErrorsE2E(t *testing.T) {
	e := newExpect(t)

	e.POST("/api/v1/auth/password-reset").
		WithJSON(map[string]any{"password": "new-password"}).
		Expect().
		Status(http.StatusUnprocessableEntity)

	e.POST("/api/v1/auth/password-reset").
		WithJSON(map[string]any{"token": "some-token", "password": "short"}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func TestListSessionsRequiresTokenE2E(t *testing.T) {
	newExpect(t).GET("/api/v1/auth/sessions").
		Expect().
		Status(http.StatusUnauthorized)
}

func TestListSessionsE2E(t *testing.T) {
	user := registerAndLogin(t)

	sessions := newExpect(t).GET("/api/v1/auth/sessions").
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("sessions").
		Array()

	sessions.Length().Gt(0)
	sessions.Element(0).Object().Keys().Contains("id", "user_id", "device_name", "ip_address", "expires_at")
}

func TestRevokeOwnSessionE2E(t *testing.T) {
	user := registerAndLogin(t)

	e := newExpect(t)

	sessionID := firstSessionID(t, user.tokens.AccessToken)

	e.DELETE("/api/v1/auth/sessions/"+strconv.FormatUint(sessionID, 10)).
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusNoContent)

	e.POST("/api/v1/auth/refresh").
		WithJSON(map[string]any{"refresh_token": user.tokens.RefreshToken}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRevokeOtherUserSessionForbiddenE2E(t *testing.T) {
	userA := registerAndLogin(t)
	userB := registerAndLogin(t)

	sessionID := firstSessionID(t, userA.tokens.AccessToken)

	newExpect(t).DELETE("/api/v1/auth/sessions/"+strconv.FormatUint(sessionID, 10)).
		WithHeader("Authorization", "Bearer "+userB.tokens.AccessToken).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestRevokeMissingSessionE2E(t *testing.T) {
	user := registerAndLogin(t)

	newExpect(t).DELETE("/api/v1/auth/sessions/999999").
		WithHeader("Authorization", "Bearer "+user.tokens.AccessToken).
		Expect().
		Status(http.StatusNotFound)
}

func TestRevokeSessionRequiresTokenE2E(t *testing.T) {
	newExpect(t).DELETE("/api/v1/auth/sessions/1").
		Expect().
		Status(http.StatusUnauthorized)
}

func firstSessionID(t *testing.T, accessToken string) uint64 {
	t.Helper()

	sessions := newExpect(t).GET("/api/v1/auth/sessions").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("sessions").
		Array()

	return uint64(sessions.Element(0).Object().Value("id").Number().Raw())
}
