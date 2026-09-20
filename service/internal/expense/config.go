package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/systemconfig"
)

const (
	configExpenseJournalID         = "expense.journal_id"
	configExpenseEmployeePayable   = "expense.employee_payable_account_id"
	configExpenseCardClearing      = "expense.card_clearing_account_id"
	configExpenseReimbursementBank = "expense.reimbursement_bank_account_id"
)

type ExpenseConfigSource interface {
	JournalID(ctx context.Context, organizationID uint64) (uint64, error)
	EmployeePayableAccountID(ctx context.Context, organizationID uint64) (uint64, error)
	CardClearingAccountID(ctx context.Context, organizationID uint64) (uint64, error)
	ReimbursementBankAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type expenseConfigSource struct {
	configs systemconfig.Lookup
}

func NewExpenseConfigSource(configs systemconfig.Lookup) ExpenseConfigSource {
	return expenseConfigSource{configs: configs}
}

func (s expenseConfigSource) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configExpenseJournalID, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (s expenseConfigSource) EmployeePayableAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configExpenseEmployeePayable, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (s expenseConfigSource) CardClearingAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configExpenseCardClearing, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (s expenseConfigSource) ReimbursementBankAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	var value uint64
	if err := systemconfig.Read(ctx, s.configs, organizationID, configExpenseReimbursementBank, &value); err != nil {
		return 0, err
	}
	return value, nil
}
