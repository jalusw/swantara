//go:build e2e

package e2e

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

const (
	adminEmail    = "admin@swantara.id"
	adminPassword = "password123"
)

func login(t *testing.T, email, password string) authTokens {
	t.Helper()

	loginData := newExpect(t).POST("/api/v1/auth/login").
		WithJSON(map[string]any{"email": email, "password": password}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object()

	return authTokens{
		AccessToken:  loginData.Value("access_token").String().Raw(),
		RefreshToken: loginData.Value("refresh_token").String().Raw(),
	}
}

func createOrganization(t *testing.T, accessToken string) uint64 {
	t.Helper()

	org := newExpect(t).POST("/api/v1/organizations").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(map[string]any{
			"name":          "E2E " + gofakeit.Company(),
			"base_currency": "IDR",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("organization").
		Object()

	return uint64(org.Value("id").Number().Raw())
}

func TestTenantResourcesNestedOnlyE2E(t *testing.T) {
	admin := login(t, adminEmail, adminPassword)
	orgID := createOrganization(t, admin.AccessToken)
	orgPath := "/api/v1/organizations/" + strconv.FormatUint(orgID, 10)

	e := newExpect(t)

	for _, path := range []string{"/tax-periods", "/account-movements", "/invoices", "/contacts", "/products"} {
		e.GET(orgPath+path).
			WithHeader("Authorization", "Bearer "+admin.AccessToken).
			Expect().
			Status(http.StatusOK)
	}
}

func TestRemovedTopLevelTenantRoutesE2E(t *testing.T) {
	admin := login(t, adminEmail, adminPassword)

	e := newExpect(t)

	for _, path := range []string{
		"/api/v1/tax-periods",
		"/api/v1/tax-rules",
		"/api/v1/account-movements",
		"/api/v1/invoices",
		"/api/v1/payments",
		"/api/v1/budgets",
		"/api/v1/contacts",
		"/api/v1/products",
		"/api/v1/warehouses",
	} {
		e.GET(path).
			WithHeader("Authorization", "Bearer "+admin.AccessToken).
			Expect().
			Status(http.StatusNotFound)
	}
}

func TestGlobalRoutesStillAvailableE2E(t *testing.T) {
	admin := login(t, adminEmail, adminPassword)

	newExpect(t).GET("/api/v1/units").
		WithHeader("Authorization", "Bearer "+admin.AccessToken).
		Expect().
		Status(http.StatusOK)
}

func TestTenantResourceDeniedForNonMemberE2E(t *testing.T) {
	admin := login(t, adminEmail, adminPassword)
	orgID := createOrganization(t, admin.AccessToken)

	outsider := registerAndLogin(t)

	newExpect(t).GET("/api/v1/organizations/"+strconv.FormatUint(orgID, 10)+"/tax-periods").
		WithHeader("Authorization", "Bearer "+outsider.tokens.AccessToken).
		Expect().
		Status(http.StatusForbidden)
}
