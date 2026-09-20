package seeders

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestSeeder_New(t *testing.T) {
	passwordSvc := iam.PasswordService{}
	if _, err := New(nil, passwordSvc); !helper.AssertError(t, err, true, ErrDatabaseNotInitialized) {
		return
	}
	db, _ := query.NewMockDB(t)
	if _, err := New(db, passwordSvc); err != nil {
		t.Fatalf("New(db) error = %v", err)
	}
}

func TestSeedMissing_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	repo := dao.NewBase[reference.Currency](db)

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "currencies" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created, err := seedMissing(context.Background(), repo, "code", nil, []*reference.Currency{{Code: "IDR"}}, func(c *reference.Currency) any { return c.Code })
	if err != nil {
		t.Fatalf("seedMissing error = %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_SkipsExisting(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	repo := dao.NewBase[reference.Currency](db)

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created, err := seedMissing(context.Background(), repo, "code", nil, []*reference.Currency{{Code: "IDR"}}, func(c *reference.Currency) any { return c.Code })
	if err != nil {
		t.Fatalf("seedMissing error = %v", err)
	}
	if created {
		t.Error("expected created = false")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	repo := dao.NewBase[reference.Currency](db)

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnError(errors.New("db down"))

	_, err := seedMissing(context.Background(), repo, "code", nil, []*reference.Currency{{Code: "IDR"}}, func(c *reference.Currency) any { return c.Code })
	if err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	repo := dao.NewBase[reference.Currency](db)

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "currencies" .*`).
		WillReturnError(errors.New("db down"))

	_, err := seedMissing(context.Background(), repo, "code", nil, []*reference.Currency{{Code: "IDR"}}, func(c *reference.Currency) any { return c.Code })
	if err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_WithOrganization_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	repo := dao.NewBase[reference.FxRate](db)
	orgID := uint64(1)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "fx_rates" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "fx_rates" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "fx_rates" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created, err := seedMissing(context.Background(), repo, "currency_code", &orgID, []*reference.FxRate{{CurrencyCode: "USD"}}, func(r *reference.FxRate) any { return r.CurrencyCode })
	if err != nil {
		t.Fatalf("seedMissing error = %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_WithOrganization_SkipsExisting(t *testing.T) {
	db, mock := query.NewMockDB(t)
	repo := dao.NewBase[reference.FxRate](db)
	orgID := uint64(1)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "fx_rates" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "fx_rates" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created, err := seedMissing(context.Background(), repo, "currency_code", &orgID, []*reference.FxRate{{CurrencyCode: "USD"}}, func(r *reference.FxRate) any { return r.CurrencyCode })
	if err != nil {
		t.Fatalf("seedMissing error = %v", err)
	}
	if created {
		t.Error("expected created = false")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedMissing_WithOrganization_ReturnsListError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	repo := dao.NewBase[reference.FxRate](db)
	orgID := uint64(1)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "fx_rates" WHERE .*`).
		WillReturnError(errors.New("db down"))

	_, err := seedMissing(context.Background(), repo, "currency_code", &orgID, []*reference.FxRate{{CurrencyCode: "USD"}}, func(r *reference.FxRate) any { return r.CurrencyCode })
	if err == nil {
		t.Fatal("expected list error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestRefID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	repo := dao.NewBase[reference.Currency](db)

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	id, err := refID(context.Background(), repo, "code", "IDR")
	if err != nil {
		t.Fatalf("refID error = %v", err)
	}
	if id != 42 {
		t.Errorf("refID = %d, want 42", id)
	}

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if _, err := refID(context.Background(), repo, "code", "USD"); !helper.AssertError(t, err, true, ErrReferenceMissing) {
		return
	}

	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnError(errors.New("db down"))
	if _, err := refID(context.Background(), repo, "code", "EUR"); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedFoundation_AllExisting(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	accounting, err := loadAccountingSeed()
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := loadInventorySeed()
	if err != nil {
		t.Fatal(err)
	}
	hr, err := loadHRSeed()
	if err != nil {
		t.Fatal(err)
	}
	crm, err := loadCRMSeed()
	if err != nil {
		t.Fatal(err)
	}
	manufacturing, err := loadManufacturingSeed()
	if err != nil {
		t.Fatal(err)
	}
	pos, err := loadPOSSeed()
	if err != nil {
		t.Fatal(err)
	}
	subscription, err := loadSubscriptionSeed()
	if err != nil {
		t.Fatal(err)
	}
	system, err := loadSystemSeed()
	if err != nil {
		t.Fatal(err)
	}
	currencySeedData, err := LoadCurrencySeed()
	if err != nil {
		t.Fatal(err)
	}

	journalRefIDs := 0
	for _, journal := range accounting.Journals {
		if journal.DefaultAccount != "" {
			journalRefIDs++
		}
	}
	categoryRefIDs := 0
	for _, category := range inventory.ProductCategories {
		for _, code := range []string{
			category.IncomeAccount,
			category.ExpenseAccount,
			category.CogsAccount,
			category.StockValuationAccount,
			category.StockInputAccount,
			category.StockOutputAccount,
		} {
			if code != "" {
				categoryRefIDs++
			}
		}
	}

	queries := 1 + // defaultOrg
		len(currencySeedData.Currencies) + // seedCurrencies
		2*len(accounting.FxRates) + // seedFxRates
		len(inventory.UnitCategories) + len(inventory.Units) + // seedUnits
		2*len(accounting.PaymentTerms) + // seedPaymentTerms
		2*len(accounting.Accounts) + // seedAccounts
		len(accounting.Dimensions) + // seedDimensions
		journalRefIDs + 2*len(accounting.Journals) + // seedJournals
		2 + // seedTaxYear
		categoryRefIDs + 2*len(inventory.ProductCategories) + len(inventory.ItemAttributes) + // seedProductCatalog
		2*len(crm.PipelineStages) + 2*len(crm.SalesGroups) + // seedCRM
		len(inventory.Carriers) + // seedCarriers
		len(accounting.ReminderLevels) + // seedReminder
		1 + len(inventory.StockLocations) + // seedWarehouses
		len(hr.Departments) + len(hr.JobPositions) + // seedDepartments
		3*len(manufacturing.WorkCenters) + // seedWorkCenters
		len(hr.LeaveTypes) + // seedLeaveTypes
		3*len(hr.SalaryRules) + // seedSalaryRules
		7 + // seedAssetCategories
		3 + // seedQualityPoints
		4*len(pos.POSConfigs) + // seedPOSConfigs
		3*len(accounting.ExpenseCategories) + // seedExpenseCategories
		2*len(subscription.SubscriptionPlans) + // seedSubscriptionPlans
		2*len(system.SystemConfigs) + // seedSystemConfigs
		2*len(system.DocSequences) // seedDocSequences

	physicalLocations := 0
	for _, loc := range inventory.StockLocations {
		if loc.Physical {
			physicalLocations++
		}
	}

	mock.MatchExpectationsInOrder(false)
	for i := 0; i < queries+100; i++ {
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < physicalLocations; i++ {
		mock.ExpectBegin()
		mock.ExpectExec(".*").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	if err := seeder.SeedFoundation(); err != nil {
		t.Fatalf("SeedFoundation error = %v", err)
	}
}

func TestLoadSeedData_ReturnsReadError(t *testing.T) {
	if _, err := loadSeedData[any]("data/missing.json"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestSeeder_defaultOrg_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "organizations" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	orgID, err := seeder.defaultOrg(context.Background())
	if err != nil {
		t.Fatalf("defaultOrg error = %v", err)
	}
	if orgID != 7 {
		t.Errorf("defaultOrg = %d, want 7", orgID)
	}
}

func TestSeeder_seedApproverIDs(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "system_configs" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "system_configs" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "system_configs" .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	if err := seeder.seedApproverIDs(context.Background(), 1, []uint64{2}); err != nil {
		t.Fatalf("seedApproverIDs error = %v", err)
	}
}

func TestSeeder_SeedDemoUsers_AlreadySeeded(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	if err := seeder.SeedDemoUsers(); err != nil {
		t.Fatalf("SeedDemoUsers error = %v", err)
	}
}

func TestSeeder_SeedDemoUsers_CreatesAdmin(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(".*").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	for i := 0; i < 700; i++ {
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.SeedDemoUsers(); err != nil {
		t.Fatalf("SeedDemoUsers error = %v", err)
	}
}
