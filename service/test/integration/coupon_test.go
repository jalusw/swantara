//go:build integration

package integration

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestCoupon_RedemptionHonorsLimitsAndAudits(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	svc := giftcard.NewCouponService(giftcard.NewCouponDAO(testDB), db.NewDBTransactioner(testDB))

	limit := 1
	coupon, err := svc.Create(ctx, &giftcard.Coupon{
		OrganizationID: helper.Ptr(fx.orgID),
		Code:           "SAVE10",
		DiscountType:   giftcard.CouponDiscountPercent,
		DiscountValue:  10,
		UsageLimit:     &limit,
	})
	if err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	result, err := svc.Redeem(ctx, giftcard.RedeemCouponRequest{
		OrganizationID: fx.orgID,
		Code:           "SAVE10",
		Subtotal:       amount.FromFloat64(200),
	})
	if err != nil {
		t.Fatalf("redeem coupon failed: %v", err)
	}
	if result.Coupon.UsedCount != 1 || !result.Discount.Equal(amount.FromFloat64(20)) {
		t.Fatalf("redeem = used_count %d, discount %v; want 1 and 20", result.Coupon.UsedCount, result.Discount)
	}

	if _, err := svc.Redeem(ctx, giftcard.RedeemCouponRequest{
		OrganizationID: fx.orgID,
		Code:           "SAVE10",
		Subtotal:       amount.FromFloat64(200),
	}); !errors.Is(err, giftcard.ErrCouponOverLimit) {
		t.Fatalf("second redeem err = %v, want ErrCouponOverLimit", err)
	}

	expired := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.Create(ctx, &giftcard.Coupon{
		OrganizationID: helper.Ptr(fx.orgID),
		Code:           "GONE10",
		DiscountType:   giftcard.CouponDiscountFixed,
		DiscountValue:  10,
		ExpiryDate:     &expired,
	}); err != nil {
		t.Fatalf("create expired coupon failed: %v", err)
	}
	if _, err := svc.Validate(ctx, fx.orgID, "GONE10", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, giftcard.ErrCouponExpired) {
		t.Fatalf("validate expired coupon err = %v, want ErrCouponExpired", err)
	}

	page, err := giftcard.NewCouponDAO(testDB).List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "code", Operator: query.Equal, Value: coupon.Code}},
	})
	if err != nil {
		t.Fatalf("list coupons failed: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("coupons = %d, want 1", len(page.Items))
	}

	var auditRows []struct {
		Diff string
	}
	if err := testDB.Raw(`SELECT diff FROM audit_logs WHERE table_name = 'coupons' AND record_id = ? AND action = 'insert'`, coupon.ID).Scan(&auditRows).Error; err != nil {
		t.Fatalf("query audit log failed: %v", err)
	}
	if len(auditRows) != 1 {
		t.Fatalf("audit rows = %d, want 1", len(auditRows))
	}
	var diff map[string]any
	if err := json.Unmarshal([]byte(auditRows[0].Diff), &diff); err != nil {
		t.Fatalf("unmarshal diff failed: %v", err)
	}
	if diff["code"] != "[REDACTED]" {
		t.Errorf("coupon code in audit diff = %#v, want redacted", diff["code"])
	}
}
