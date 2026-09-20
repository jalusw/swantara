package quality

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestQualityPointDAO_ListByItem(t *testing.T) {
	db, mock := query.NewMockDB(t)
	pointDAO := NewQualityPointDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quality_points" .*`).
		WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "quality_points" .*`).
		WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "test_type", "norm_min", "norm_max", "operation", "created_at", "updated_at"}).
			AddRow(1, 2, TestTypePassFail, nil, nil, nil, now, now))

	points, err := pointDAO.ListByItem(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 1 || points[0].ID != 1 || points[0].ItemID == nil {
		t.Errorf("points = %+v", points)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityCheckDAO_ListByShipment(t *testing.T) {
	db, mock := query.NewMockDB(t)
	checkDAO := NewQualityCheckDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quality_checks" .*`).
		WithArgs(uint64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "quality_checks" .*`).
		WithArgs(uint64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "point_id", "item_id", "shipment_id", "result", "created_at", "updated_at"}).
			AddRow(1, 1, 2, 3, CheckResultPending, now, now))

	checks, err := checkDAO.ListByShipment(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(checks) != 1 || checks[0].Result != CheckResultPending {
		t.Errorf("checks = %+v", checks)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityCheckDAO_ListInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	checkDAO := NewQualityCheckDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	joinPattern := `JOIN item_variants ON item_variants\.id = quality_checks\.item_id JOIN items ON items\.id = item_variants\.item_id AND items\.organization_id`
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quality_checks" ` + joinPattern + ` .*`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT "quality_checks"\."id".* FROM "quality_checks" ` + joinPattern + ` .*`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "result", "created_at", "updated_at"}).
			AddRow(1, CheckResultPass, now, now))

	page, err := checkDAO.ListInOrg(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 || page.Items[0].ID != 1 {
		t.Errorf("page = %+v", page)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityCheckDAO_FindInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	checkDAO := NewQualityCheckDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT "quality_checks"\."id".* FROM "quality_checks" JOIN item_variants ON .* WHERE quality_checks\.id = \$2 .*`).
		WithArgs(uint64(10), uint64(7), uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "result", "created_at", "updated_at"}).
			AddRow(7, CheckResultPass, now, now))

	check, err := checkDAO.FindInOrg(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check == nil || check.ID != 7 {
		t.Errorf("check = %+v", check)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityCheckDAO_FindInOrg_NotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	checkDAO := NewQualityCheckDAO(db)

	mock.ExpectQuery(`SELECT "quality_checks"\."id".* FROM "quality_checks" JOIN item_variants ON .*`).
		WithArgs(uint64(10), uint64(7), uint64(1)).
		WillReturnError(gorm.ErrRecordNotFound)

	check, err := checkDAO.FindInOrg(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check != nil {
		t.Errorf("check = %+v, want nil", check)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityAlertDAO_ListByCheck(t *testing.T) {
	db, mock := query.NewMockDB(t)
	alertDAO := NewQualityAlertDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quality_alerts" .*`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "quality_alerts" .*`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "check_id", "state", "created_at", "updated_at"}).
			AddRow(1, 1, AlertStateOpen, now, now))

	alerts, err := alertDAO.ListByCheck(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 || alerts[0].State != AlertStateOpen {
		t.Errorf("alerts = %+v", alerts)
	}
	query.AssertDBMockDone(t, mock)
}

func TestQualityAlertDAO_ListInOrgAndFindInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	alertDAO := NewQualityAlertDAO(db)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	joinPattern := `JOIN item_variants ON item_variants\.id = quality_alerts\.item_id JOIN items ON items\.id = item_variants\.item_id AND items\.organization_id`
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quality_alerts" ` + joinPattern + ` .*`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT "quality_alerts"\."id".* FROM "quality_alerts" ` + joinPattern + ` .*`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state", "created_at", "updated_at"}).
			AddRow(1, AlertStateOpen, now, now))

	page, err := alertDAO.ListInOrg(context.Background(), nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("page = %+v", page)
	}

	mock.ExpectQuery(`SELECT "quality_alerts"\."id".* FROM "quality_alerts" JOIN item_variants ON .* WHERE quality_alerts\.id = \$2 .*`).
		WithArgs(uint64(10), uint64(7), uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state", "created_at", "updated_at"}).
			AddRow(7, AlertStateSolved, now, now))

	alert, err := alertDAO.FindInOrg(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert == nil || alert.ID != 7 || alert.State != AlertStateSolved {
		t.Errorf("alert = %+v", alert)
	}
	query.AssertDBMockDone(t, mock)
}
