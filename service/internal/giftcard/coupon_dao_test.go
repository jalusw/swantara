package giftcard

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestCouponDAO_RedeemTx_IncrementsAndReturnsCoupon(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "coupons" SET "used_count"=used_count + 1 WHERE id = $1 AND (usage_limit IS NULL OR used_count < usage_limit) AND deleted_at IS NULL`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "coupons" WHERE id = $1 LIMIT $2`)).
		WithArgs(42, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "used_count"}).AddRow(42, "SAVE10", 1))

	coupons := NewCouponDAO(db)

	coupon, err := coupons.RedeemTx(ctx, db, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coupon.ID != 42 || coupon.UsedCount != 1 {
		t.Errorf("coupon = %+v, want id 42 used once", coupon)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCouponDAO_RedeemTx_RejectsOverLimit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "coupons" SET "used_count"=used_count + 1`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	coupons := NewCouponDAO(db)

	_, err := coupons.RedeemTx(ctx, db, 42)
	if helper.AssertError(t, err, true, ErrCouponOverLimit) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCouponDAO_RedeemTx_PropagatesUpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "coupons" SET "used_count"=used_count + 1`)).
		WithArgs(42).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	coupons := NewCouponDAO(db)

	_, err := coupons.RedeemTx(ctx, db, 42)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCouponDAO_RedeemTx_PropagatesReadError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "coupons" SET "used_count"=used_count + 1`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "coupons" WHERE id = $1`)).
		WithArgs(42, 1).
		WillReturnError(errors.New("db down"))

	coupons := NewCouponDAO(db)

	_, err := coupons.RedeemTx(ctx, db, 42)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewCouponDAO_Constructs(t *testing.T) {
	db, _ := query.NewMockDB(t)
	coupons := NewCouponDAO(db)
	if coupons == nil {
		t.Fatal("expected non-nil dao")
	}
}
