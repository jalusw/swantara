package procurement

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

var currencyRateColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "from_currency", "to_currency", "rate", "rate_date"}

func currencyRateRow(rate float64) *sqlmock.Rows {
	return sqlmock.NewRows(currencyRateColumns).AddRow(1, time.Now(), time.Now(), nil, 10, "USD", "IDR", rate, time.Now())
}

func TestCurrencyRateDAO_FindLatest(t *testing.T) {
	t.Run("finds latest by pair", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currency_rates"`)).
			WillReturnRows(currencyRateRow(16000))

		got, err := NewCurrencyRateDAO(db).FindByPair(context.Background(), "USD", "IDR", helper.Ptr(uint64(10)))
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Rate != 16000 {
			t.Errorf("rate = %v, want 16000", got.Rate)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("finds by pair and date without org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currency_rates"`)).
			WillReturnRows(currencyRateRow(15500))

		got, err := NewCurrencyRateDAO(db).FindByPairDate(context.Background(), "USD", "IDR", nil, time.Now())
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Rate != 15500 {
			t.Errorf("rate = %v, want 15500", got.Rate)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when missing", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currency_rates"`)).
			WillReturnRows(sqlmock.NewRows(currencyRateColumns))

		got, err := NewCurrencyRateDAO(db).FindByPair(context.Background(), "USD", "IDR", nil)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("rate = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currency_rates"`)).WillReturnError(dbErr)

		_, err := NewCurrencyRateDAO(db).FindByPair(context.Background(), "USD", "IDR", nil)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestSupplyAgreementDAO_CreateWithLines(t *testing.T) {
	t.Run("creates agreement and lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supply_agreements"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supply_agreement_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectCommit()

		agreement := &SupplyAgreement{SupplierID: 10}
		lines := []*SupplyAgreementLine{{Qty: 2, UnitPrice: 1000}}
		created, err := NewSupplyAgreementDAO(db).CreateWithLines(context.Background(), agreement, lines)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if created.ID != 1 || lines[0].AgreementID != 1 {
			t.Errorf("created = %+v line = %+v, want linked ids", created, lines[0])
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rolls back on line error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supply_agreements"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supply_agreement_lines"`)).
			WillReturnError(errors.New("line down"))
		mock.ExpectRollback()

		_, err := NewSupplyAgreementDAO(db).CreateWithLines(context.Background(), &SupplyAgreement{SupplierID: 10}, []*SupplyAgreementLine{{Qty: 2}})
		if err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})
}

func TestSupplyAgreementLineDAO_ListByAgreement(t *testing.T) {
	t.Run("returns lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supply_agreement_lines" WHERE agreement_id = $1`)).
			WithArgs(7).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supply_agreement_lines" WHERE agreement_id = $1`)).
			WithArgs(7).
			WillReturnRows(sqlmock.NewRows([]string{"id", "agreement_id", "qty"}).AddRow(1, 7, 3))

		items, err := NewSupplyAgreementLineDAO(db).ListByAgreement(context.Background(), 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 || items[0].Qty != 3 {
			t.Errorf("items = %+v, want one line", items)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supply_agreement_lines"`)).WillReturnError(dbErr)

		_, err := NewSupplyAgreementLineDAO(db).ListByAgreement(context.Background(), 7)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

var scorecardColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "supplier_id", "period_start", "period_end", "quality_score", "delivery_score", "price_score", "overall_score", "total_orders", "on_time_deliveries", "quality_failures"}

func scorecardRow() *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(scorecardColumns).AddRow(1, now, now, nil, 10, 20, now, now, 90, 85, 80, 85, 10, 9, 1)
}

func TestSupplierScorecardDAO_FindByVendorAndPeriod(t *testing.T) {
	now := time.Now()

	t.Run("finds with period bounds", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_scorecards"`)).
			WillReturnRows(scorecardRow())

		got, err := NewSupplierScorecardDAO(db).FindByVendorAndPeriod(context.Background(), 20, &now, &now)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.OverallScore != 85 {
			t.Errorf("score = %v, want 85", got.OverallScore)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("finds without period bounds", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_scorecards"`)).
			WillReturnRows(scorecardRow())

		got, err := NewSupplierScorecardDAO(db).FindByVendorAndPeriod(context.Background(), 20, nil, nil)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.SupplierID != 20 {
			t.Errorf("supplier = %d, want 20", got.SupplierID)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when missing", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_scorecards"`)).
			WillReturnRows(sqlmock.NewRows(scorecardColumns))

		got, err := NewSupplierScorecardDAO(db).FindByVendorAndPeriod(context.Background(), 20, nil, nil)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("scorecard = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_scorecards"`)).WillReturnError(dbErr)

		_, err := NewSupplierScorecardDAO(db).FindByVendorAndPeriod(context.Background(), 20, nil, nil)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestPaymentBatchDAO_CreateWithLines(t *testing.T) {
	t.Run("creates batch and lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_batches"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_batch_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectCommit()

		batch := &PaymentBatch{ContactID: 20}
		lines := []*PaymentBatchLine{{OrderID: 7, Amount: 5000}}
		created, err := NewPaymentBatchDAO(db).CreateWithLines(context.Background(), batch, lines)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if created.ID != 1 || lines[0].BatchID != 1 {
			t.Errorf("created = %+v line = %+v, want linked ids", created, lines[0])
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rolls back on batch error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_batches"`)).
			WillReturnError(errors.New("batch down"))
		mock.ExpectRollback()

		_, err := NewPaymentBatchDAO(db).CreateWithLines(context.Background(), &PaymentBatch{}, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})
}

func TestPaymentBatchLineDAO_ListByBatch(t *testing.T) {
	t.Run("returns lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_batch_lines" WHERE batch_id = $1`)).
			WithArgs(7).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_batch_lines" WHERE batch_id = $1`)).
			WithArgs(7).
			WillReturnRows(sqlmock.NewRows([]string{"id", "batch_id", "amount"}).AddRow(1, 7, 5000))

		items, err := NewPaymentBatchLineDAO(db).ListByBatch(context.Background(), 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 || items[0].Amount != 5000 {
			t.Errorf("items = %+v, want one line", items)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_batch_lines"`)).WillReturnError(dbErr)

		_, err := NewPaymentBatchLineDAO(db).ListByBatch(context.Background(), 7)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestProcurementDAO_Constructors(t *testing.T) {
	db, _ := query.NewMockDB(t)
	constructors := map[string]any{
		"currency rates":  NewCurrencyRateDAO(db),
		"agreements":      NewSupplyAgreementDAO(db),
		"agreement lines": NewSupplyAgreementLineDAO(db),
		"scorecards":      NewSupplierScorecardDAO(db),
		"cost centers":    NewCostCenterDAO(db),
		"credit memos":    NewPurchaseCreditMemoDAO(db),
		"debit memos":     NewPurchaseDebitMemoDAO(db),
		"batches":         NewPaymentBatchDAO(db),
		"batch lines":     NewPaymentBatchLineDAO(db),
	}
	for name, dao := range constructors {
		if dao == nil {
			t.Errorf("%s dao = nil", name)
		}
	}
}
