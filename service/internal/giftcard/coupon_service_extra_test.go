package giftcard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestCouponService_Create_PersistsCoupon(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	var created *Coupon
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return nil, nil },
			CreateFunc: func(_ context.Context, coupon *Coupon) (*Coupon, error) {
				created = coupon
				return coupon, nil
			},
		},
	}, TransactionerMock{})

	coupon := &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10}
	got, err := svc.Create(ctx, coupon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil {
		t.Error("coupon was not persisted")
	}
	if got != coupon {
		t.Errorf("got = %+v, want original coupon", got)
	}
}

func TestCouponService_Create_RejectsNegativeUsageLimit(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	negative := -1
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10, UsageLimit: &negative})
	if helper.AssertError(t, err, true, ErrCouponDiscountInvalid) {
		return
	}
}

func TestCouponService_Create_PropagatesSearchError(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return nil, errors.New("db down") },
		},
	}, TransactionerMock{})

	_, err := svc.Create(ctx, &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestCouponService_Update_PersistsChanges(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	coupon := &Coupon{Base: model.Base{ID: 42}, OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 15}
	var updated *Coupon
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return nil, nil },
			UpdateFunc: func(_ context.Context, coupon *Coupon) (*Coupon, error) {
				updated = coupon
				return coupon, nil
			},
		},
	}, TransactionerMock{})

	got, err := svc.Update(ctx, coupon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil {
		t.Error("coupon was not persisted")
	}
	if got != coupon {
		t.Errorf("got = %+v, want original coupon", got)
	}
}

func TestCouponService_Update_RejectsInvalidCoupon(t *testing.T) {
	ctx := context.Background()
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Update(ctx, &Coupon{Code: "SAVE10"})
	if helper.AssertError(t, err, true, ErrCouponOrganization) {
		return
	}
}

func TestCouponService_Validate_RejectsEmptyCode(t *testing.T) {
	ctx := context.Background()
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Validate(ctx, 10, "", time.Now())
	if helper.AssertError(t, err, true, ErrCouponCodeRequired) {
		return
	}
}

func TestCouponService_Validate_RejectsUnknownCode(t *testing.T) {
	ctx := context.Background()
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Validate(ctx, 10, "NOPE", time.Now())
	if helper.AssertError(t, err, true, ErrCouponNotFound) {
		return
	}
}

func TestCouponService_Validate_RejectsOtherOrganization(t *testing.T) {
	ctx := context.Background()
	org := uint64(20)
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) {
				return &Coupon{OrganizationID: &org, Code: "SAVE10"}, nil
			},
		},
	}, TransactionerMock{})

	_, err := svc.Validate(ctx, 10, "SAVE10", time.Now())
	if helper.AssertError(t, err, true, ErrCouponNotFound) {
		return
	}
}

func TestCouponService_Validate_PropagatesSearchError(t *testing.T) {
	ctx := context.Background()
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return nil, errors.New("db down") },
		},
	}, TransactionerMock{})

	_, err := svc.Validate(ctx, 10, "SAVE10", time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestCouponService_Discount_RejectsNegativeSubtotal(t *testing.T) {
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	got := svc.Discount(&Coupon{DiscountType: CouponDiscountPercent, DiscountValue: 10}, amount.FromFloat64(-50))
	if !got.IsZero() {
		t.Errorf("discount = %v, want zero", got)
	}
}

func TestCouponService_Redeem_RequiresOrganization(t *testing.T) {
	ctx := context.Background()
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Redeem(ctx, RedeemCouponRequest{Code: "SAVE10", Subtotal: amount.FromFloat64(200)})
	if helper.AssertError(t, err, true, ErrCouponOrganization) {
		return
	}
}

func TestCouponService_Redeem_PropagatesValidationError(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})

	_, err := svc.Redeem(ctx, RedeemCouponRequest{OrganizationID: org, Code: "NOPE", Subtotal: amount.FromFloat64(200)})
	if helper.AssertError(t, err, true, ErrCouponNotFound) {
		return
	}
}

func TestCouponService_Redeem_UsesDefaultRedeem(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	coupon := &Coupon{Base: model.Base{ID: 42}, OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10}
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return coupon, nil },
		},
	}, TransactionerMock{})

	result, err := svc.Redeem(ctx, RedeemCouponRequest{OrganizationID: org, Code: "SAVE10", Subtotal: amount.FromFloat64(200)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Discount.Equal(amount.FromFloat64(20)) {
		t.Errorf("discount = %v, want 20", result.Discount)
	}
}
