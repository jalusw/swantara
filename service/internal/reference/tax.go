package reference

import "github.com/jalusw/swantara/apps/service/internal/kernel/amount"

func TaxLineAmount(subtotal, qty amount.Amount, tax *Tax) (amount.Amount, error) {
	if tax.Amount == nil {
		return amount.Zero(), nil
	}
	switch tax.Type {
	case TaxTypePercent:
		rate := amount.FromFloat64(*tax.Amount / 100)
		return subtotal.Mul(rate).Round(4), nil
	case TaxTypeFixed:
		return qty.Mul(amount.FromFloat64(*tax.Amount)).Round(4), nil
	default:
		return amount.Zero(), nil
	}
}
