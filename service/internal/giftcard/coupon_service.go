package giftcard

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type CouponService struct {
	coupons CouponDAO
	tx      db.Transactioner
	now     func() time.Time
}

func NewCouponService(coupons CouponDAO, tx db.Transactioner) CouponService {
	return CouponService{coupons: coupons, tx: tx, now: time.Now}
}

func (s CouponService) List(ctx context.Context, q *query.Query) (*query.Page[Coupon], error) {
	return s.coupons.List(ctx, q)
}

func (s CouponService) Find(ctx context.Context, id uint64) (*Coupon, error) {
	return s.coupons.Find(ctx, id)
}

func (s CouponService) Delete(ctx context.Context, id uint64) error {
	return s.coupons.Delete(ctx, id)
}

func (s CouponService) Create(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	if err := s.validate(ctx, coupon, 0); err != nil {
		return nil, err
	}
	return s.coupons.Create(ctx, coupon)
}

func (s CouponService) Update(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	if err := s.validate(ctx, coupon, coupon.ID); err != nil {
		return nil, err
	}
	return s.coupons.Update(ctx, coupon)
}

func (s CouponService) validate(ctx context.Context, coupon *Coupon, excludeID uint64) error {
	if coupon.OrganizationID == nil {
		return ErrCouponOrganization
	}
	if coupon.Code == "" {
		return ErrCouponCodeRequired
	}
	if _, ok := validCouponDiscountType[coupon.DiscountType]; !ok {
		return ErrCouponDiscountInvalid
	}
	if coupon.DiscountValue <= 0 {
		return ErrCouponDiscountInvalid
	}
	if coupon.UsageLimit != nil && *coupon.UsageLimit < 0 {
		return ErrCouponDiscountInvalid
	}

	existing, err := s.coupons.Search(ctx, "code", coupon.Code)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludeID && existing.OrganizationID != nil && *existing.OrganizationID == *coupon.OrganizationID {
		return ErrCouponAlreadyExists
	}
	return nil
}

func (s CouponService) Validate(ctx context.Context, organizationID uint64, code string, date time.Time) (*Coupon, error) {
	if code == "" {
		return nil, ErrCouponCodeRequired
	}
	coupon, err := s.coupons.Search(ctx, "code", code)
	if err != nil {
		return nil, err
	}
	if coupon == nil || coupon.OrganizationID == nil || *coupon.OrganizationID != organizationID {
		return nil, ErrCouponNotFound
	}
	if err := s.assertRedeemable(coupon, date); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (s CouponService) assertRedeemable(coupon *Coupon, date time.Time) error {
	if coupon.ExpiryDate != nil && !date.IsZero() && date.After(*coupon.ExpiryDate) {
		return ErrCouponExpired
	}
	if coupon.UsageLimit != nil && coupon.UsedCount >= *coupon.UsageLimit {
		return ErrCouponOverLimit
	}
	return nil
}

func (s CouponService) Discount(coupon *Coupon, subtotal amount.Amount) amount.Amount {
	if subtotal.IsNegative() {
		return amount.Zero()
	}
	discount := amount.Zero()
	switch coupon.DiscountType {
	case CouponDiscountPercent:
		discount = subtotal.Mul(amount.FromFloat64(coupon.DiscountValue / 100)).Round(4)
	case CouponDiscountFixed:
		discount = amount.FromFloat64(coupon.DiscountValue).Round(4)
		if discount.GreaterThan(subtotal) {
			discount = subtotal
		}
	}
	if discount.IsNegative() {
		return amount.Zero()
	}
	return discount
}

type RedeemCouponRequest struct {
	OrganizationID uint64
	Code           string
	Subtotal       amount.Amount
	Date           time.Time
}

type RedeemCouponResult struct {
	Coupon   *Coupon       `json:"coupon"`
	Discount amount.Amount `json:"discount"`
}

func (s CouponService) Redeem(ctx context.Context, request RedeemCouponRequest) (RedeemCouponResult, error) {
	if request.OrganizationID == 0 {
		return RedeemCouponResult{}, ErrCouponOrganization
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	coupon, err := s.Validate(ctx, request.OrganizationID, request.Code, date)
	if err != nil {
		return RedeemCouponResult{}, err
	}
	discount := s.Discount(coupon, request.Subtotal)

	var redeemed *Coupon
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		redeemed, err = s.coupons.RedeemTx(ctx, tx, coupon.ID)
		return err
	})
	if err != nil {
		return RedeemCouponResult{}, err
	}
	return RedeemCouponResult{Coupon: redeemed, Discount: discount}, nil
}
