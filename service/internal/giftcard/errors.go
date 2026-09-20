package giftcard

import "errors"

var (
	ErrGiftCardNotFound       = errors.New("gift card not found")
	ErrGiftCardCodeRequired   = errors.New("gift card code is required")
	ErrGiftCardAmountInvalid  = errors.New("gift card amount must be positive")
	ErrGiftCardInactive       = errors.New("gift card is not active")
	ErrGiftCardExpired        = errors.New("gift card has expired")
	ErrGiftCardInsufficient   = errors.New("gift card balance is insufficient")
	ErrGiftCardAccounts       = errors.New("gift card requires cash and liability accounts")
	ErrGiftCardLiabilityOnly  = errors.New("gift card requires the liability account")
	ErrGiftCardRevenue        = errors.New("gift card redemption requires a revenue account")
	ErrGiftCardRefundAccount  = errors.New("gift card refund requires a revenue refund account")
	ErrGiftCardAlreadyExists  = errors.New("gift card code already exists")
	ErrGiftCardForfeitAccount = errors.New("gift card forfeiture requires liability and income accounts")
	ErrGiftCardOrganization   = errors.New("gift card forfeiture requires an organization")

	ErrCouponNotFound        = errors.New("coupon not found")
	ErrCouponCodeRequired    = errors.New("coupon code is required")
	ErrCouponOrganization    = errors.New("coupon requires an organization")
	ErrCouponDiscountInvalid = errors.New("coupon discount is invalid")
	ErrCouponAlreadyExists   = errors.New("coupon code already exists for organization")
	ErrCouponExpired         = errors.New("coupon has expired")
	ErrCouponOverLimit       = errors.New("coupon usage limit exceeded")
	ErrCouponInactive        = errors.New("coupon is not active")
)
