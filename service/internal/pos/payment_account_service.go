package pos

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type PaymentAccountService struct {
	accounts dao.CRUD[reference.POSPaymentAccount]
}

func NewPaymentAccountService(accounts dao.CRUD[reference.POSPaymentAccount]) *PaymentAccountService {
	return &PaymentAccountService{accounts: accounts}
}

func (s *PaymentAccountService) List(ctx context.Context, q *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
	return s.accounts.List(ctx, q)
}

func (s *PaymentAccountService) Upsert(ctx context.Context, organizationID uint64, method string, accountID uint64) (*reference.POSPaymentAccount, bool, error) {
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "organization_id", Operator: query.Equal, Value: organizationID},
		{Field: "method", Operator: query.Equal, Value: method},
	}})
	if err != nil {
		return nil, false, err
	}

	if len(page.Items) > 0 {
		existing := page.Items[0]
		existing.AccountID = accountID
		updated, err := s.accounts.Update(ctx, existing)
		if err != nil {
			return nil, false, err
		}
		return updated, false, nil
	}

	created, err := s.accounts.Create(ctx, &reference.POSPaymentAccount{
		OrganizationID: organizationID,
		Method:         method,
		AccountID:      accountID,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, false, ErrPaymentAccountAlreadyExists
		}
		return nil, false, err
	}
	return created, true, nil
}

func (s *PaymentAccountService) Delete(ctx context.Context, organizationID uint64, method string) error {
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "organization_id", Operator: query.Equal, Value: organizationID},
		{Field: "method", Operator: query.Equal, Value: method},
	}})
	if err != nil {
		return err
	}
	if len(page.Items) == 0 {
		return ErrPaymentAccountNotFound
	}
	return s.accounts.Delete(ctx, page.Items[0].ID)
}
