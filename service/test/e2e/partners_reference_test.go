//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestContactRelationsLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	address := f.authed(t, http.MethodPost, "/contacts/"+itoa(f.contactID)+"/addresses").
		WithJSON(map[string]any{
			"label":  "HQ",
			"street": gofakeit.Street(),
			"city":   gofakeit.City(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("address").
		Object()

	addressID := uint64(address.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/contacts/"+itoa(f.contactID)+"/addresses/"+itoa(addressID)+"/default").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/contacts/"+itoa(f.contactID)+"/bank-accounts").
		WithJSON(map[string]any{
			"account_number": gofakeit.UUID(),
			"bank_name":      "E2E Bank",
		}).
		Expect().
		Status(http.StatusCreated)
}

func TestReferenceCatalogE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	admin := login(t, adminEmail, adminPassword)

	e.GET("/api/v1/units").
		WithHeader("Authorization", "Bearer "+admin.AccessToken).
		Expect().
		Status(http.StatusOK)

	e.GET("/api/v1/currencies").
		WithHeader("Authorization", "Bearer "+admin.AccessToken).
		Expect().
		Status(http.StatusOK)

	for _, path := range []string{"/accounts", "/journals", "/taxes", "/tax-years", "/dimension-accounts"} {
		f.authed(t, http.MethodGet, path).
			Expect().
			Status(http.StatusOK)
	}
}

func TestOrganizationLifecycleE2E(t *testing.T) {
	admin := login(t, adminEmail, adminPassword)
	e := newExpect(t)

	orgID := createOrganization(t, admin.AccessToken)

	e.GET("/api/v1/organizations/"+itoa(orgID)).
		WithHeader("Authorization", "Bearer "+admin.AccessToken).
		Expect().
		Status(http.StatusOK)
}
