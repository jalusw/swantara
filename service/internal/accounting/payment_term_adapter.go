package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type referencePaymentTermAdapter struct {
	terms reference.PaymentTermDAO
}

func NewPaymentTermAdapter(terms reference.PaymentTermDAO) PaymentTermSplitter {
	return referencePaymentTermAdapter{terms: terms}
}

func (a referencePaymentTermAdapter) Splits(ctx context.Context, termID uint64, total amount.Amount, date time.Time) ([]PaymentTermSplit, error) {
	svc := reference.NewPaymentTermService(a.terms)
	splits, err := svc.Splits(ctx, termID, total, date)
	if err != nil {
		return nil, err
	}
	out := make([]PaymentTermSplit, len(splits))
	for i, split := range splits {
		out[i] = PaymentTermSplit{Amount: split.Amount, DueDate: split.DueDate, Sequence: split.Sequence}
	}
	return out, nil
}
