package pos

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func paymentAccountMock(items []*reference.POSPaymentAccount, listErr error) dao.CRUDMock[reference.POSPaymentAccount] {
	return dao.CRUDMock[reference.POSPaymentAccount]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
			return &query.Page[reference.POSPaymentAccount]{Items: items, Count: int64(len(items))}, listErr
		},
	}
}

func TestPaymentAccountService(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("lists", func(t *testing.T) {
		svc := NewPaymentAccountService(paymentAccountMock(nil, nil))
		if svc == nil {
			t.Fatal("service = nil")
		}
		page, err := svc.List(ctx, &query.Query{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 0 {
			t.Errorf("count = %d", page.Count)
		}
	})

	t.Run("upserts existing", func(t *testing.T) {
		existing := &reference.POSPaymentAccount{Base: model.Base{ID: 1}, OrganizationID: 10, Method: "cash", AccountID: 100}
		mock := paymentAccountMock([]*reference.POSPaymentAccount{existing}, nil)
		mock.UpdateFunc = func(_ context.Context, a *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
			return a, nil
		}
		svc := NewPaymentAccountService(mock)
		got, created, err := svc.Upsert(ctx, 10, "cash", 200)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if created || got.AccountID != 200 {
			t.Errorf("upsert = %+v created %v", got, created)
		}
	})

	t.Run("creates new", func(t *testing.T) {
		mock := paymentAccountMock(nil, nil)
		mock.CreateFunc = func(_ context.Context, a *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
			a.ID = 2
			return a, nil
		}
		svc := NewPaymentAccountService(mock)
		got, created, err := svc.Upsert(ctx, 10, "cash", 200)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if !created || got.ID != 2 {
			t.Errorf("upsert = %+v created %v", got, created)
		}
	})

	t.Run("upsert failures", func(t *testing.T) {
		svc := NewPaymentAccountService(paymentAccountMock(nil, dbErr))
		_, _, err := svc.Upsert(ctx, 10, "cash", 200)
		helper.AssertError(t, err, true, dbErr)

		updating := paymentAccountMock([]*reference.POSPaymentAccount{{Base: model.Base{ID: 1}}}, nil)
		updating.UpdateFunc = func(_ context.Context, _ *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
			return nil, dbErr
		}
		svc = NewPaymentAccountService(updating)
		_, _, err = svc.Upsert(ctx, 10, "cash", 200)
		helper.AssertError(t, err, true, dbErr)

		conflicting := paymentAccountMock(nil, nil)
		conflicting.CreateFunc = func(_ context.Context, _ *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		}
		svc = NewPaymentAccountService(conflicting)
		_, _, err = svc.Upsert(ctx, 10, "cash", 200)
		helper.AssertError(t, err, true, ErrPaymentAccountAlreadyExists)

		failing := paymentAccountMock(nil, nil)
		failing.CreateFunc = func(_ context.Context, _ *reference.POSPaymentAccount) (*reference.POSPaymentAccount, error) {
			return nil, dbErr
		}
		svc = NewPaymentAccountService(failing)
		_, _, err = svc.Upsert(ctx, 10, "cash", 200)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("deletes", func(t *testing.T) {
		var deleted uint64
		mock := paymentAccountMock([]*reference.POSPaymentAccount{{Base: model.Base{ID: 1}}}, nil)
		mock.DeleteFunc = func(_ context.Context, id uint64) error {
			deleted = id
			return nil
		}
		svc := NewPaymentAccountService(mock)
		if err := svc.Delete(ctx, 10, "cash"); helper.AssertError(t, err, false, nil) {
			return
		}
		if deleted != 1 {
			t.Errorf("deleted = %d", deleted)
		}
	})

	t.Run("delete failures", func(t *testing.T) {
		svc := NewPaymentAccountService(paymentAccountMock(nil, dbErr))
		helper.AssertError(t, svc.Delete(ctx, 10, "cash"), true, dbErr)

		svc = NewPaymentAccountService(paymentAccountMock(nil, nil))
		helper.AssertError(t, svc.Delete(ctx, 10, "cash"), true, ErrPaymentAccountNotFound)

		failing := paymentAccountMock([]*reference.POSPaymentAccount{{Base: model.Base{ID: 1}}}, nil)
		failing.DeleteFunc = func(_ context.Context, _ uint64) error { return dbErr }
		svc = NewPaymentAccountService(failing)
		helper.AssertError(t, svc.Delete(ctx, 10, "cash"), true, dbErr)
	})
}

func TestPOSConfigService_CRUD(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	svc := NewPOSConfigService(dao.CRUDMock[reference.POSConfig]{})
	if _, err := svc.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := svc.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if _, err := svc.Create(ctx, &reference.POSConfig{}); err != nil {
		t.Errorf("Create = %v", err)
	}
	if _, err := svc.Update(ctx, &reference.POSConfig{}); err != nil {
		t.Errorf("Update = %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
}

func TestPOSFixtures(t *testing.T) {
	if POSSessionFixture() == nil {
		t.Error("session = nil")
	}
	if POSOrderFixture() == nil {
		t.Error("order = nil")
	}
	if POSOrderLineFixture() == nil {
		t.Error("line = nil")
	}
	if POSPaymentFixture() == nil {
		t.Error("payment = nil")
	}
}

func posCountSelect(mock sqlmock.Sqlmock, table string) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
}

func TestPOSDAO_Queries(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("session lists in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_sessions" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "pos_sessions" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		page, err := NewPOSSessionDAO(db).ListInOrganization(ctx, nil, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d", page.Count)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("session list error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_sessions" JOIN`)).WillReturnError(dbErr)
		_, err := NewPOSSessionDAO(db).ListInOrganization(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("order lists in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_orders" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "pos_orders" JOIN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		page, err := NewPOSOrderDAO(db).ListInOrganization(ctx, nil, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d", page.Count)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("order list error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_orders" JOIN`)).WillReturnError(dbErr)
		_, err := NewPOSOrderDAO(db).ListInOrganization(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("order lines and payments by order", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		posCountSelect(mock, "pos_order_lines")
		items, err := NewPOSOrderLineDAO(db).ListByOrder(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		posCountSelect(mock, "pos_payments")
		payments, err := NewPOSPaymentDAO(db).ListByOrder(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(payments) != 1 {
			t.Errorf("len = %d", len(payments))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("sums by session", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "pos_payments"`)).
			WillReturnRows(sqlmock.NewRows([]string{"method", "total"}).AddRow("cash", 1500.0))
		totals, err := NewPOSPaymentDAO(db).SumBySession(ctx, 3)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(totals) != 1 || totals[0].Total != 1500 {
			t.Errorf("totals = %+v", totals)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("sum error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "pos_payments"`)).WillReturnError(dbErr)
		_, err := NewPOSPaymentDAO(db).SumBySession(ctx, 3)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("update tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pos_orders"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewPOSOrderDAO(db).UpdateTx(ctx, tx, &POSOrder{Base: model.Base{ID: 1}})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("creates with lines and payments", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_order_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_payments"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
		mock.ExpectCommit()

		order := &POSOrder{}
		lines := []*POSOrderLine{{Qty: 1}}
		payments := []*POSPayment{{Amount: 100}}
		got, err := NewPOSOrderDAO(db).CreateWithLinesAndPayments(ctx, order, lines, payments)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 || lines[0].OrderID != 1 || payments[0].OrderID != 1 {
			t.Errorf("order = %+v", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("create with lines rolls back", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_order_lines"`)).WillReturnError(dbErr)
		mock.ExpectRollback()

		_, err := NewPOSOrderDAO(db).CreateWithLinesAndPayments(ctx, &POSOrder{}, []*POSOrderLine{{}}, nil)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("constructors", func(t *testing.T) {
		db, _ := query.NewMockDB(t)
		if NewPOSSessionDAO(db) == nil {
			t.Error("sessions = nil")
		}
		if NewPOSOrderDAO(db) == nil {
			t.Error("orders = nil")
		}
		if NewPOSOrderLineDAO(db) == nil {
			t.Error("lines = nil")
		}
		if NewPOSPaymentDAO(db) == nil {
			t.Error("payments = nil")
		}
	})
}

func TestPOSMock_Fallbacks(t *testing.T) {
	ctx := context.Background()

	t.Run("transactioner", func(t *testing.T) {
		if err := (TransactionerMock{}).Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
			t.Errorf("Run = %v", err)
		}
	})

	t.Run("session dao", func(t *testing.T) {
		if _, err := (POSSessionDAOMock{}).ListInOrganization(ctx, &query.Query{}, 10); err != nil {
			t.Errorf("ListInOrganization = %v", err)
		}
		withFunc := POSSessionDAOMock{
			ListInOrganizationFunc: func(_ context.Context, _ *query.Query, _ uint64) (*query.Page[POSSession], error) {
				return &query.Page[POSSession]{}, nil
			},
		}
		if _, err := withFunc.ListInOrganization(ctx, &query.Query{}, 10); err != nil {
			t.Errorf("ListInOrganization = %v", err)
		}
	})

	t.Run("order dao", func(t *testing.T) {
		mock := POSOrderDAOMock{}
		if _, err := mock.CreateWithLinesAndPayments(ctx, &POSOrder{}, nil, nil); err != nil {
			t.Errorf("CreateWithLinesAndPayments = %v", err)
		}
		if _, err := mock.ListInOrganization(ctx, &query.Query{}, 10); err != nil {
			t.Errorf("ListInOrganization = %v", err)
		}
		withFunc := POSOrderDAOMock{
			CreateWithLinesAndPaymentsFunc: func(_ context.Context, o *POSOrder, _ []*POSOrderLine, _ []*POSPayment) (*POSOrder, error) {
				return o, nil
			},
			ListInOrganizationFunc: func(_ context.Context, _ *query.Query, _ uint64) (*query.Page[POSOrder], error) {
				return &query.Page[POSOrder]{}, nil
			},
		}
		if _, err := withFunc.CreateWithLinesAndPayments(ctx, &POSOrder{}, nil, nil); err != nil {
			t.Errorf("CreateWithLinesAndPayments = %v", err)
		}
		if _, err := withFunc.ListInOrganization(ctx, &query.Query{}, 10); err != nil {
			t.Errorf("ListInOrganization = %v", err)
		}
	})

	t.Run("line and payment dao", func(t *testing.T) {
		if _, err := (POSOrderLineDAOMock{}).ListByOrder(ctx, 1); err != nil {
			t.Errorf("ListByOrder = %v", err)
		}
		if _, err := (POSPaymentDAOMock{}).ListByOrder(ctx, 1); err != nil {
			t.Errorf("ListByOrder = %v", err)
		}
		withFunc := POSPaymentDAOMock{
			ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSPayment, error) {
				return []*POSPayment{{}}, nil
			},
		}
		if _, err := withFunc.ListByOrder(ctx, 1); err != nil {
			t.Errorf("ListByOrder = %v", err)
		}
	})
}

var _ = driver.Value(nil)
