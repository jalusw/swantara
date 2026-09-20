package giftcard

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type CouponDAO interface {
	dao.CRUD[Coupon]
	RedeemTx(ctx context.Context, tx *gorm.DB, couponID uint64) (*Coupon, error)
}

type couponDAO struct {
	dao.Base[Coupon]
	db *gorm.DB
}

func NewCouponDAO(db *gorm.DB) CouponDAO {
	return couponDAO{Base: dao.NewBase[Coupon](db), db: db}
}

func (d couponDAO) RedeemTx(ctx context.Context, tx *gorm.DB, couponID uint64) (*Coupon, error) {
	result := tx.WithContext(ctx).
		Model(&Coupon{}).
		Where("id = ? AND (usage_limit IS NULL OR used_count < usage_limit) AND deleted_at IS NULL", couponID).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrCouponOverLimit
	}
	var coupon Coupon
	if err := tx.WithContext(ctx).Where("id = ?", couponID).Take(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}
