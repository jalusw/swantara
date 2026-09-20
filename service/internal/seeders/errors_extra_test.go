package seeders

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func failingSeeder(t *testing.T) *Seeder {
	t.Helper()
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))
	mock.ExpectExec(".*").WillReturnError(errors.New("db down"))
	svc, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestSeedFunctions_PropagateErrors(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		call func(svc *Seeder) error
	}{
		{name: "currencies", call: func(svc *Seeder) error { return svc.seedCurrencies(ctx) }},
		{name: "fx rates", call: func(svc *Seeder) error { return svc.seedFxRates(ctx, 1) }},
		{name: "units", call: func(svc *Seeder) error { return svc.seedUnits(ctx) }},
		{name: "payment terms", call: func(svc *Seeder) error { return svc.seedPaymentTerms(ctx) }},
		{name: "accounts", call: func(svc *Seeder) error { return svc.seedAccounts(ctx, 1) }},
		{name: "dimension accounts", call: func(svc *Seeder) error { return svc.seedDimensions(ctx, 1) }},
		{name: "journals", call: func(svc *Seeder) error { return svc.seedJournals(ctx, 1) }},
		{name: "tax year", call: func(svc *Seeder) error { return svc.seedTaxYear(ctx, 1) }},
		{name: "item catalog", call: func(svc *Seeder) error { return svc.seedProductCatalog(ctx, 1) }},
		{name: "crm", call: func(svc *Seeder) error { return svc.seedCRM(ctx, 1) }},
		{name: "carriers", call: func(svc *Seeder) error { return svc.seedCarriers(ctx) }},
		{name: "reminder", call: func(svc *Seeder) error { return svc.seedReminder(ctx) }},
		{name: "warehouses", call: func(svc *Seeder) error { return svc.seedWarehouses(ctx, 1) }},
		{name: "departments", call: func(svc *Seeder) error { return svc.seedDepartments(ctx, 1) }},
		{name: "work centers", call: func(svc *Seeder) error { return svc.seedWorkCenters(ctx, 1) }},
		{name: "leave types", call: func(svc *Seeder) error { return svc.seedLeaveTypes(ctx) }},
		{name: "salary rules", call: func(svc *Seeder) error { return svc.seedSalaryRules(ctx) }},
		{name: "asset categories", call: func(svc *Seeder) error { return svc.seedAssetCategories(ctx, 1) }},
		{name: "quality points", call: func(svc *Seeder) error { return svc.seedQualityPoints(ctx, 1) }},
		{name: "pos configs", call: func(svc *Seeder) error { return svc.seedPOSConfigs(ctx, 1) }},
		{name: "expense categories", call: func(svc *Seeder) error { return svc.seedExpenseCategories(ctx, 1) }},
		{name: "subscription plans", call: func(svc *Seeder) error { return svc.seedSubscriptionPlans(ctx, 1) }},
		{name: "system configs", call: func(svc *Seeder) error { return svc.seedSystemConfigs(ctx, 1) }},
		{name: "doc sequences", call: func(svc *Seeder) error { return svc.seedDocSequences(ctx, 1) }},
		{name: "approver ids", call: func(svc *Seeder) error { return svc.seedApproverIDs(ctx, 1, []uint64{7}) }},
		{name: "foundation", call: func(svc *Seeder) error { return svc.SeedFoundation() }},
		{name: "demo users", call: func(svc *Seeder) error { return svc.SeedDemoUsers() }},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(failingSeeder(t)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
