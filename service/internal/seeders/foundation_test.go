package seeders

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestSeeder_defaultOrg_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if _, err := seeder.defaultOrg(context.Background()); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_defaultOrg_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "organizations" .*`).
		WillReturnError(errors.New("db down"))

	if _, err := seeder.defaultOrg(context.Background()); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedJournals_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedJournals(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedWorkCenters_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "dimensions" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedWorkCenters(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedSalaryRules_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedSalaryRules(context.Background()); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedAssetCategories_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedAssetCategories(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedQualityPoints_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedQualityPoints(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPOSConfigs_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedPOSConfigs(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedExpenseCategories_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedExpenseCategories(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedCRM_ReturnsListError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT count\(\*\) FROM "pipeline_stages" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedCRM(context.Background(), 1); err == nil {
		t.Fatal("expected list error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDimensions_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "dimensions" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDimensions(context.Background(), 1); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDimensions_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "dimensions" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`INSERT INTO "dimensions" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedDimensions(context.Background(), 1); err != nil {
		t.Fatalf("seedDimensions error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDepartments(context.Background(), 1); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`INSERT INTO "departments" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "job_positions" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`INSERT INTO "job_positions" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedDepartments(context.Background(), 1); err != nil {
		t.Fatalf("seedDepartments error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_ReturnsJobPositionRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "job_positions" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDepartments(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_ReturnsJobPositionCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "job_positions" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO "job_positions" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDepartments(context.Background(), 1); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedUnits_ReturnsCategoryCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "unit_groups" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedUnits(context.Background()); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedUnits_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	seed, err := loadInventorySeed()
	if err != nil {
		t.Fatalf("loadInventorySeed error = %v", err)
	}

	for i := 0; i < len(seed.UnitCategories); i++ {
		mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedUnits(context.Background()); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedUnits_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	seed, err := loadInventorySeed()
	if err != nil {
		t.Fatalf("loadInventorySeed error = %v", err)
	}

	for i := 0; i < len(seed.UnitCategories); i++ {
		mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedUnits(context.Background()); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedUnits_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	seed, err := loadInventorySeed()
	if err != nil {
		t.Fatalf("loadInventorySeed error = %v", err)
	}

	for i := 0; i < len(seed.UnitCategories); i++ {
		mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO "units" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedUnits(context.Background()); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedUnits_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	seed, err := loadInventorySeed()
	if err != nil {
		t.Fatalf("loadInventorySeed error = %v", err)
	}

	for i := 0; i < len(seed.UnitCategories); i++ {
		mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < len(seed.Units); i++ {
		mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`SELECT \* FROM "unit_groups" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`INSERT INTO "units" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedUnits(context.Background()); err != nil {
		t.Fatalf("seedUnits error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPaymentTerms_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.MatchExpectationsInOrder(false)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Swantara Demo"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payment_terms" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedPaymentTerms(context.Background()); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPaymentTerms_ReturnsListLinesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Swantara Demo"))
	mock.ExpectQuery(`.*`).WillReturnError(errors.New("db down"))

	if err := seeder.seedPaymentTerms(context.Background()); err == nil {
		t.Fatal("expected list error")
	}
}

func TestSeeder_seedPaymentTerms_ReturnsReplaceLinesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Swantara Demo"))
	mock.ExpectQuery(`.*`).WillReturnError(errors.New("db down"))

	if err := seeder.seedPaymentTerms(context.Background()); err == nil {
		t.Fatal("expected replace error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPaymentTerms_CreatesTermsAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.MatchExpectationsInOrder(false)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Swantara Demo"))
	for i := 0; i < 22; i++ {
		mock.ExpectQuery(`.*`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`.*`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedPaymentTerms(context.Background()); err != nil {
		t.Fatalf("seedPaymentTerms error = %v", err)
	}

}

func TestSeeder_seedWarehouses_ReturnsSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedWarehouses(context.Background(), 1); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedWarehouses_CreatesWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "warehouses" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	for i := 0; i < 10; i++ {
		mock.ExpectQuery(`SELECT \* FROM "stock_locations" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`INSERT INTO "stock_locations" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedWarehouses(context.Background(), 1); err != nil {
		t.Fatalf("seedWarehouses error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedProductCatalog_ReturnsRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedProductCatalog(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedProductCatalog_CreatesAttributeValues(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 2; i++ {
		mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 10; i++ {
		mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 3; i++ {
		mock.ExpectQuery(`SELECT count\(\*\) FROM "item_categories" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`SELECT \* FROM "item_categories" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`INSERT INTO "item_categories" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 3; i++ {
		mock.ExpectQuery(`SELECT \* FROM "items" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`SELECT \* FROM "item_categories" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`SELECT \* FROM "units" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`INSERT INTO "items" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`INSERT INTO "item_variants" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 2; i++ {
		mock.ExpectQuery(`SELECT \* FROM "item_attributes" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(`INSERT INTO "item_attributes" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	for i := 0; i < 8; i++ {
		mock.ExpectQuery(`INSERT INTO "item_attribute_values" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := seeder.seedProductCatalog(context.Background(), 1); err != nil {
		t.Fatalf("seedProductCatalog error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPaymentTerms_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.MatchExpectationsInOrder(false)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Swantara Demo"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payment_terms" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "payment_terms" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "payment_terms" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedPaymentTerms(context.Background()); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDimensions_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "dimensions" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "dimensions" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDimensions(context.Background(), 1); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "departments" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDepartments(context.Background(), 1); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedDepartments_ReturnsJobPositionSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT \* FROM "departments" WHERE .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	mock.ExpectQuery(`SELECT \* FROM "job_positions" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedDepartments(context.Background(), 1); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedWarehouses_ReturnsCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "warehouses" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedWarehouses(context.Background(), 1); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedWarehouses_ReturnsLocationSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "stock_locations" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedWarehouses(context.Background(), 1); err == nil {
		t.Fatal("expected search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedSalaryRules_ReturnsCreditRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedSalaryRules(context.Background()); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedPOSConfigs_ReturnsJournalRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "journals" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedPOSConfigs(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedWarehouses_ReturnsLocationCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "warehouses" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "stock_locations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO "stock_locations" .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedWarehouses(context.Background(), 1); err == nil {
		t.Fatal("expected create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_seedAssetCategories_ReturnsDepreciationRefError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.seedAssetCategories(context.Background(), 1); err == nil {
		t.Fatal("expected reference lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedFoundation_ReturnsCurrenciesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(`SELECT \* FROM "organizations" WHERE .*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "currencies" WHERE .*`).
		WillReturnError(errors.New("db down"))

	if err := seeder.SeedFoundation(); err == nil {
		t.Fatal("expected currencies error")
	}
	query.AssertDBMockDone(t, mock)
}
