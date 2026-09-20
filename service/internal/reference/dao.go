package reference

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type AccountDAO interface {
	List(ctx context.Context, q *query.Query) (*query.Page[Account], error)
	Create(ctx context.Context, entity *Account) (*Account, error)
}

type TaxDAO interface {
	List(ctx context.Context, q *query.Query) (*query.Page[Tax], error)
	Create(ctx context.Context, entity *Tax) (*Tax, error)
	Update(ctx context.Context, entity *Tax) (*Tax, error)
}

type WithholdingTaxDAO interface {
	List(ctx context.Context, q *query.Query) (*query.Page[WithholdingTax], error)
	Create(ctx context.Context, entity *WithholdingTax) (*WithholdingTax, error)
}

type PaymentTermDAO interface {
	dao.CRUD[PaymentTerm]
	ReplaceLines(ctx context.Context, termID uint64, lines []*PaymentTermLine) error
	DeleteWithLines(ctx context.Context, termID uint64) error
	ListLines(ctx context.Context, termID uint64) ([]*PaymentTermLine, error)
}

type paymentTermDAO struct {
	dao.Base[PaymentTerm]
	lines dao.Base[PaymentTermLine]
}

func NewPaymentTermDAO(db *gorm.DB) PaymentTermDAO {
	return paymentTermDAO{
		Base:  dao.NewBase[PaymentTerm](db),
		lines: dao.NewBase[PaymentTermLine](db),
	}
}

func (d paymentTermDAO) ListLines(ctx context.Context, termID uint64) ([]*PaymentTermLine, error) {
	page, err := d.lines.List(ctx, &query.Query{Filters: []query.Filter{{Field: "payment_term_id", Operator: query.Equal, Value: termID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d paymentTermDAO) ReplaceLines(ctx context.Context, termID uint64, lines []*PaymentTermLine) error {
	existing, err := d.ListLines(ctx, termID)
	if err != nil {
		return err
	}
	for _, line := range existing {
		if err := d.lines.HardDelete(ctx, line.ID); err != nil {
			return err
		}
	}
	for _, line := range lines {
		line.PaymentTermID = termID
		if _, err := d.lines.Create(ctx, line); err != nil {
			return err
		}
	}
	return nil
}

func (d paymentTermDAO) DeleteWithLines(ctx context.Context, termID uint64) error {
	existing, err := d.ListLines(ctx, termID)
	if err != nil {
		return err
	}
	for _, line := range existing {
		if err := d.lines.HardDelete(ctx, line.ID); err != nil {
			return err
		}
	}
	return d.Delete(ctx, termID)
}
