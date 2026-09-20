package giftcard

import "testing"

func TestGiftCardModels_TableNames(t *testing.T) {
	if (GiftCard{}).TableName() != "gift_cards" {
		t.Errorf("gift card table = %s, want gift_cards", (GiftCard{}).TableName())
	}
	if (GiftCardTransaction{}).TableName() != "gift_card_transactions" {
		t.Errorf("transaction table = %s, want gift_card_transactions", (GiftCardTransaction{}).TableName())
	}
	if (Coupon{}).TableName() != "coupons" {
		t.Errorf("coupon table = %s, want coupons", (Coupon{}).TableName())
	}
}
