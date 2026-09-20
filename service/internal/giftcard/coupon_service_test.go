package giftcard

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestCouponService_Create_ValidatesCoupon(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	tests := []struct {
		name    string
		coupon  *Coupon
		wantErr error
	}{
		{
			name:    "rejects missing organization",
			coupon:  &Coupon{Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10},
			wantErr: ErrCouponOrganization,
		},
		{
			name:    "rejects missing code",
			coupon:  &Coupon{OrganizationID: &org, DiscountType: CouponDiscountPercent, DiscountValue: 10},
			wantErr: ErrCouponCodeRequired,
		},
		{
			name:    "rejects invalid discount type",
			coupon:  &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: "bogus", DiscountValue: 10},
			wantErr: ErrCouponDiscountInvalid,
		},
		{
			name:    "rejects non-positive discount value",
			coupon:  &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 0},
			wantErr: ErrCouponDiscountInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})
			_, err := svc.Create(ctx, tt.coupon)
			if helper.AssertError(t, err, true, tt.wantErr) {
				return
			}
		})
	}
}

func TestCouponService_Create_RejectsDuplicateCode(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) {
				return &Coupon{Base: model.Base{ID: 9}, OrganizationID: &org, Code: "SAVE10"}, nil
			},
		},
	}, TransactionerMock{})

	_, err := svc.Create(ctx, &Coupon{OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10})
	if helper.AssertError(t, err, true, ErrCouponAlreadyExists) {
		return
	}
}

func TestCouponService_Validate_RejectsExpiredAndOverLimit(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	past := now.AddDate(0, 0, -1)
	limit := 5

	tests := []struct {
		name    string
		coupon  *Coupon
		wantErr error
	}{
		{
			name:    "rejects expired",
			coupon:  &Coupon{OrganizationID: &org, ExpiryDate: &past},
			wantErr: ErrCouponExpired,
		},
		{
			name:    "rejects over limit",
			coupon:  &Coupon{OrganizationID: &org, UsageLimit: &limit, UsedCount: limit},
			wantErr: ErrCouponOverLimit,
		},
		{
			name:    "accepts valid coupon",
			coupon:  &Coupon{OrganizationID: &org, UsageLimit: &limit, UsedCount: 2},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCouponService(CouponDAOMock{
				CRUDMock: dao.CRUDMock[Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return tt.coupon, nil },
				},
			}, TransactionerMock{})
			_, err := svc.Validate(ctx, org, "SAVE10", now)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
		})
	}
}

func TestCouponService_Discount_AppliesPercentAndFixed(t *testing.T) {
	subtotal := amount.FromFloat64(200)
	tests := []struct {
		name       string
		coupon     *Coupon
		wantAmount float64
	}{
		{name: "percent discount", coupon: &Coupon{DiscountType: CouponDiscountPercent, DiscountValue: 10}, wantAmount: 20},
		{name: "fixed discount", coupon: &Coupon{DiscountType: CouponDiscountFixed, DiscountValue: 50}, wantAmount: 50},
		{name: "fixed clamps to subtotal", coupon: &Coupon{DiscountType: CouponDiscountFixed, DiscountValue: 300}, wantAmount: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCouponService(CouponDAOMock{}, TransactionerMock{})
			got := svc.Discount(tt.coupon, subtotal)
			if !got.Equal(amount.FromFloat64(tt.wantAmount)) {
				t.Errorf("discount = %v, want %v", got, tt.wantAmount)
			}
		})
	}
}

func TestCouponService_Redeem_AtomicallyIncrementsUsage(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	coupon := &Coupon{Base: model.Base{ID: 42}, OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10, UsedCount: 0}
	redeemed := &Coupon{Base: model.Base{ID: 42}, OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10, UsedCount: 1}

	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return coupon, nil },
		},
		RedeemTxFunc: func(_ context.Context, _ *gorm.DB, id uint64) (*Coupon, error) {
			return redeemed, nil
		},
	}, TransactionerMock{})

	result, err := svc.Redeem(ctx, RedeemCouponRequest{OrganizationID: org, Code: "SAVE10", Subtotal: amount.FromFloat64(200)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Coupon.UsedCount != 1 {
		t.Errorf("used_count = %d, want 1", result.Coupon.UsedCount)
	}
	if !result.Discount.Equal(amount.FromFloat64(20)) {
		t.Errorf("discount = %v, want 20", result.Discount)
	}
}

func TestCouponService_Redeem_PropagatesLimitError(t *testing.T) {
	ctx := context.Background()
	org := uint64(10)
	coupon := &Coupon{Base: model.Base{ID: 42}, OrganizationID: &org, Code: "SAVE10", DiscountType: CouponDiscountPercent, DiscountValue: 10}

	svc := NewCouponService(CouponDAOMock{
		CRUDMock: dao.CRUDMock[Coupon]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Coupon, error) { return coupon, nil },
		},
		RedeemTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*Coupon, error) {
			return nil, ErrCouponOverLimit
		},
	}, TransactionerMock{})

	_, err := svc.Redeem(ctx, RedeemCouponRequest{OrganizationID: org, Code: "SAVE10", Subtotal: amount.FromFloat64(200)})
	if helper.AssertError(t, err, true, ErrCouponOverLimit) {
		return
	}
}
