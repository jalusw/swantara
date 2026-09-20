//go:build integration

package integration

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/seeders"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestCurrencies_SeededCatalogMatchesSeed(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	currencySeed, err := seeders.LoadCurrencySeed()
	if err != nil {
		t.Fatalf("load currency seed failed: %v", err)
	}

	currencies := dao.NewBase[reference.Currency](testDB)

	page, err := currencies.List(ctx, nil)
	if err != nil {
		t.Fatalf("list currencies failed: %v", err)
	}
	if page.Count != int64(len(currencySeed.Currencies)) {
		t.Fatalf("currency count = %d, want %d", page.Count, len(currencySeed.Currencies))
	}

	var want seeders.CurrencySeedItem
	found := false
	for _, item := range currencySeed.Currencies {
		if item.Code == "USD" {
			want = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("USD missing from currency seed")
	}
	usd, err := currencies.Search(ctx, "code", "USD")
	if err != nil {
		t.Fatalf("search USD failed: %v", err)
	}
	if usd == nil {
		t.Fatal("seeded USD not found")
	}
	wantRounding := math.Pow(10, -float64(want.DecimalPlaces))
	if usd.Name != want.Name || usd.DecimalPlaces != want.DecimalPlaces || usd.Rounding != wantRounding {
		t.Errorf("USD = %+v, want name %s decimals %d rounding %v", usd, want.Name, want.DecimalPlaces, wantRounding)
	}
	if usd.Symbol == nil || *usd.Symbol != want.Symbol {
		t.Errorf("USD symbol = %v, want %s", usd.Symbol, want.Symbol)
	}

	missing, err := currencies.Search(ctx, "code", "ZZZ")
	if err != nil {
		t.Fatalf("search ZZZ failed: %v", err)
	}
	if missing != nil {
		t.Errorf("ZZZ = %+v, want nil", missing)
	}
}

func TestFxRates_ValidatePersistAndResolve(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	otherOrg, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create second organization failed: %v", err)
	}

	svc := reference.NewFxRateService(dao.NewBase[reference.FxRate](testDB), dao.NewBase[reference.Currency](testDB))
	source := reference.NewFxRateSource(dao.NewBase[reference.FxRate](testDB))

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	if _, err := svc.Create(ctx, &reference.FxRate{CurrencyCode: "XXX", Rate: 1, ValidFrom: jan1}); !errors.Is(err, reference.ErrCurrencyNotFound) {
		t.Errorf("unknown currency err = %v, want ErrCurrencyNotFound", err)
	}
	if _, err := svc.Create(ctx, &reference.FxRate{CurrencyCode: "USD", Rate: 0, ValidFrom: jan1}); !errors.Is(err, amount.ErrInvalidRate) {
		t.Errorf("zero rate err = %v, want ErrInvalidRate", err)
	}
	if _, err := svc.Create(ctx, &reference.FxRate{CurrencyCode: "USD", Rate: 1, RateType: "midnight", ValidFrom: jan1}); !errors.Is(err, amount.ErrInvalidRate) {
		t.Errorf("invalid rate type err = %v, want ErrInvalidRate", err)
	}
	if _, err := svc.Create(ctx, &reference.FxRate{CurrencyCode: "USD", Rate: 1}); !errors.Is(err, reference.ErrRateValidFromMissing) {
		t.Errorf("missing valid from err = %v, want ErrRateValidFromMissing", err)
	}

	global, err := svc.Create(ctx, &reference.FxRate{CurrencyCode: "USD", Rate: 15000, ValidFrom: jan1})
	if err != nil {
		t.Fatalf("create global rate failed: %v", err)
	}
	if global.RateType != string(amount.RateSpot) {
		t.Errorf("default rate type = %q, want spot", global.RateType)
	}
	orgRate, err := svc.Create(ctx, &reference.FxRate{
		CurrencyCode: "USD", OrganizationID: helper.Ptr(org.ID), Rate: 15500, ValidFrom: jan1,
	})
	if err != nil {
		t.Fatalf("create org rate failed: %v", err)
	}

	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	rate, err := source.Rate(ctx, "USD", org.ID, amount.RateSpot, jan15)
	if err != nil {
		t.Fatalf("resolve org rate failed: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(15500)) {
		t.Errorf("org resolve = %v, want 15500", rate)
	}

	rate, err = source.Rate(ctx, "USD", otherOrg.ID, amount.RateSpot, jan15)
	if err != nil {
		t.Fatalf("resolve fallback rate failed: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(15000)) {
		t.Errorf("global fallback resolve = %v, want 15000", rate)
	}

	newerRate, err := svc.Create(ctx, &reference.FxRate{
		CurrencyCode: "USD", OrganizationID: helper.Ptr(org.ID), Rate: 16000, ValidFrom: feb1,
	})
	if err != nil {
		t.Fatalf("create newer org rate failed: %v", err)
	}
	feb15 := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	rate, err = source.Rate(ctx, "USD", org.ID, amount.RateSpot, feb15)
	if err != nil {
		t.Fatalf("resolve latest rate failed: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(16000)) {
		t.Errorf("latest resolve = %v, want 16000", rate)
	}

	rates := dao.NewBase[reference.FxRate](testDB)
	found, err := rates.Find(ctx, orgRate.ID)
	if err != nil {
		t.Fatalf("find org rate failed: %v", err)
	}
	found.Rate = 15700
	updated, err := svc.Update(ctx, found)
	if err != nil {
		t.Fatalf("update org rate failed: %v", err)
	}
	if updated.Rate != 15700 {
		t.Errorf("updated rate = %v, want 15700", updated.Rate)
	}
	found.Rate = 0
	if _, err := svc.Update(ctx, found); !errors.Is(err, amount.ErrInvalidRate) {
		t.Errorf("update zero rate err = %v, want ErrInvalidRate", err)
	}

	if err := rates.Delete(ctx, newerRate.ID); err != nil {
		t.Fatalf("delete newest rate failed: %v", err)
	}
	rate, err = source.Rate(ctx, "USD", org.ID, amount.RateSpot, feb15)
	if err != nil {
		t.Fatalf("resolve after delete failed: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(15700)) {
		t.Errorf("resolve after delete = %v, want 15700", rate)
	}

	if _, err := source.Rate(ctx, "EUR", org.ID, amount.RateSpot, jan15); !errors.Is(err, reference.ErrRateNotFound) {
		t.Errorf("unrated currency err = %v, want ErrRateNotFound", err)
	}
}
