package quality

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func qualityListPair(mock sqlmock.Sqlmock, table string, id uint64) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
}

func qualityListError(mock sqlmock.Sqlmock, table string, dbErr error) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).WillReturnError(dbErr)
}

func TestQualityDAO_Queries(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("point lists by item", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		qualityListPair(mock, "quality_points", 1)
		items, err := NewQualityPointDAO(db).ListByItem(ctx, 100)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("check lists by shipment", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		qualityListPair(mock, "quality_checks", 1)
		items, err := NewQualityCheckDAO(db).ListByShipment(ctx, 9)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("check lists in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_checks" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_checks" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		page, err := NewQualityCheckDAO(db).ListInOrg(ctx, nil, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d", page.Count)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("check org list errors", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_checks" JOIN`)).WillReturnError(dbErr)
		_, err := NewQualityCheckDAO(db).ListInOrg(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_checks" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_checks" JOIN`)).WillReturnError(dbErr)
		_, err = NewQualityCheckDAO(db).ListInOrg(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("check finds in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_checks" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		got, err := NewQualityCheckDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_checks" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		got, err = NewQualityCheckDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("check = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("alert lists by check", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		qualityListPair(mock, "quality_alerts", 1)
		items, err := NewQualityAlertDAO(db).ListByCheck(ctx, 4)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("alert lists in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_alerts" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_alerts" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		page, err := NewQualityAlertDAO(db).ListInOrg(ctx, nil, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d", page.Count)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("alert org list errors", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_alerts" JOIN`)).WillReturnError(dbErr)
		_, err := NewQualityAlertDAO(db).ListInOrg(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "quality_alerts" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_alerts" JOIN`)).WillReturnError(dbErr)
		_, err = NewQualityAlertDAO(db).ListInOrg(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("alert finds in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_alerts" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		got, err := NewQualityAlertDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "quality_alerts" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		got, err = NewQualityAlertDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("alert = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("list query errors", func(t *testing.T) {
		tables := []struct {
			table string
			call  func() error
		}{
			{table: "quality_points", call: func() error {
				db, mock := query.NewMockDB(t)
				qualityListError(mock, "quality_points", dbErr)
				_, err := NewQualityPointDAO(db).ListByItem(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "quality_checks", call: func() error {
				db, mock := query.NewMockDB(t)
				qualityListError(mock, "quality_checks", dbErr)
				_, err := NewQualityCheckDAO(db).ListByShipment(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "quality_alerts", call: func() error {
				db, mock := query.NewMockDB(t)
				qualityListError(mock, "quality_alerts", dbErr)
				_, err := NewQualityAlertDAO(db).ListByCheck(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
		}
		for _, tt := range tables {
			t.Run(tt.table, func(t *testing.T) {
				helper.AssertError(t, tt.call(), true, dbErr)
			})
		}
	})
}

func TestQualityMock_Fallbacks(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("point dao", func(t *testing.T) {
		mock := QualityPointDAOMock{}
		if _, err := mock.ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
		withFunc := QualityPointDAOMock{
			ListByItemFunc: func(_ context.Context, _ uint64) ([]*reference.QualityPoint, error) {
				return []*reference.QualityPoint{{}}, nil
			},
		}
		if _, err := withFunc.ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
	})

	t.Run("check dao", func(t *testing.T) {
		mock := QualityCheckDAOMock{}
		if _, err := mock.ListByShipment(ctx, 1); err != nil {
			t.Errorf("ListByShipment = %v", err)
		}
		if _, err := mock.ListInOrg(ctx, q, 10); err != nil {
			t.Errorf("ListInOrg = %v", err)
		}
		if _, err := mock.FindInOrg(ctx, 1, 10); err != nil {
			t.Errorf("FindInOrg = %v", err)
		}
		withFunc := QualityCheckDAOMock{
			ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*QualityCheck, error) { return []*QualityCheck{{}}, nil },
		}
		if _, err := withFunc.ListByShipment(ctx, 1); err != nil {
			t.Errorf("ListByShipment = %v", err)
		}
	})

	t.Run("alert dao", func(t *testing.T) {
		mock := QualityAlertDAOMock{}
		if _, err := mock.ListByCheck(ctx, 1); err != nil {
			t.Errorf("ListByCheck = %v", err)
		}
		if _, err := mock.ListInOrg(ctx, q, 10); err != nil {
			t.Errorf("ListInOrg = %v", err)
		}
		if _, err := mock.FindInOrg(ctx, 1, 10); err != nil {
			t.Errorf("FindInOrg = %v", err)
		}
	})
}

func TestQualityPointService_CRUD(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	svc := NewQualityPointService(QualityPointDAOMock{})
	if _, err := svc.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := svc.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if _, err := svc.Create(ctx, &reference.QualityPoint{}); err != nil {
		t.Errorf("Create = %v", err)
	}
	if _, err := svc.Update(ctx, &reference.QualityPoint{}); err != nil {
		t.Errorf("Update = %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
}

func TestQualityFixtures(t *testing.T) {
	if QualityCheckFixture() == nil {
		t.Error("check = nil")
	}
	if QualityAlertFixture() == nil {
		t.Error("alert = nil")
	}
}

var _ = driver.Value(nil)
