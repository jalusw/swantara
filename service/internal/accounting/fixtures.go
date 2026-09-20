package accounting

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func JournalEntryFixture(opts ...func(*JournalEntry) *JournalEntry) *JournalEntry {
	now := time.Now()
	entry := &JournalEntry{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 100)),
		JournalID:      uint64(gofakeit.Number(1, 10)),
		Date:           now,
		State:          EntryStateDraft,
	}
	for _, opt := range opts {
		opt(entry)
	}
	return entry
}

func JournalEntryPostedFixture(opts ...func(*JournalEntry) *JournalEntry) *JournalEntry {
	return JournalEntryFixture(append([]func(*JournalEntry) *JournalEntry{
		func(m *JournalEntry) *JournalEntry {
			m.State = EntryStatePosted
			return m
		},
	}, opts...)...)
}

func JournalLineFixture(opts ...func(*JournalLine) *JournalLine) *JournalLine {
	now := time.Now()
	name := gofakeit.Word()
	line := &JournalLine{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EntryID:   uint64(gofakeit.Number(1, 10000)),
		AccountID: uint64(gofakeit.Number(1, 100)),
		Name:      &name,
		Debit:     amount.FromFloat64(gofakeit.Float64Range(0, 1000)),
		Credit:    amount.Zero(),
		Date:      now,
	}
	if line.Debit.IsZero() {
		line.Credit = amount.FromFloat64(gofakeit.Float64Range(1, 1000))
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}

func BankStatementFixture(opts ...func(*BankStatement) *BankStatement) *BankStatement {
	now := time.Now()
	stmt := &BankStatement{
		Base:         model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		JournalID:    helper.Ptr(uint64(gofakeit.Number(1, 10))),
		Name:         helper.Ptr("Bank statement"),
		Date:         &now,
		BalanceStart: amount.FromFloat64(gofakeit.Float64Range(0, 10000)),
		BalanceEnd:   amount.FromFloat64(gofakeit.Float64Range(0, 10000)),
		State:        BankStatementStateOpen,
	}
	for _, opt := range opts {
		opt(stmt)
	}
	return stmt
}

func BankStatementLineFixture(opts ...func(*BankStatementLine) *BankStatementLine) *BankStatementLine {
	now := time.Now()
	line := &BankStatementLine{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		StatementID: uint64(gofakeit.Number(1, 10000)),
		Date:        &now,
		Amount:      amount.FromFloat64(gofakeit.Float64Range(1, 5000)),
		Reconciled:  false,
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}

func BudgetFixture(opts ...func(*Budget) *Budget) *Budget {
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	end := now.AddDate(0, 0, 30)
	budget := &Budget{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Name:           helper.Ptr(gofakeit.AppName()),
		DateStart:      &start,
		DateEnd:        &end,
		State:          BudgetStateDraft,
	}
	for _, opt := range opts {
		opt(budget)
	}
	return budget
}

func BudgetLineFixture(opts ...func(*BudgetLine) *BudgetLine) *BudgetLine {
	now := time.Now()
	line := &BudgetLine{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		BudgetID:      uint64(gofakeit.Number(1, 10000)),
		AccountID:     uint64(gofakeit.Number(1, 100)),
		PlannedAmount: gofakeit.Float64Range(100, 50000),
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}

func DeferredScheduleFixture(opts ...func(*DeferredSchedule) *DeferredSchedule) *DeferredSchedule {
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now.AddDate(0, 11, 0)
	schedule := &DeferredSchedule{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Type:           DeferredTypeDeferredRevenue,
		SourceType:     "journal_entry",
		SourceID:       uint64(gofakeit.Number(1, 10000)),
		TotalAmount:    gofakeit.Float64Range(1000, 500000),
		Method:         DeferredMethodLinear,
		DateStart:      &start,
		DateEnd:        &end,
		Periods:        12,
		State:          DeferredStateRunning,
	}
	for _, opt := range opts {
		opt(schedule)
	}
	return schedule
}

func DeferredScheduleLineFixture(opts ...func(*DeferredScheduleLine) *DeferredScheduleLine) *DeferredScheduleLine {
	now := time.Now()
	date := now.AddDate(0, 1, 0)
	line := &DeferredScheduleLine{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ScheduleID:      uint64(gofakeit.Number(1, 10000)),
		Sequence:        1,
		RecognitionDate: &date,
		Amount:          gofakeit.Float64Range(100, 5000),
		Posted:          false,
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}

func InvoiceFixture(opts ...func(*Invoice) *Invoice) *Invoice {
	now := time.Now()
	due := now.AddDate(0, 0, 30)
	invoice := &Invoice{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ContactID:      uint64(gofakeit.Number(1, 10000)),
		Type:           InvoiceTypeCustomerInvoice,
		Name:           helper.Ptr(gofakeit.Word()),
		InvoiceDate:    &now,
		DueDate:        &due,
		State:          InvoiceStateDraft,
		PaymentState:   PaymentStateNotPaid,
		AmountUntaxed:  amount.FromFloat64(gofakeit.Float64Range(100, 5000)),
		AmountTax:      amount.FromFloat64(gofakeit.Float64Range(0, 500)),
		AmountTotal:    amount.FromFloat64(gofakeit.Float64Range(100, 5500)),
		AmountResidual: amount.FromFloat64(gofakeit.Float64Range(100, 5500)),
	}
	for _, opt := range opts {
		opt(invoice)
	}
	return invoice
}

func PaymentFixture(opts ...func(*Payment) *Payment) *Payment {
	now := time.Now()
	payment := &Payment{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ContactID:      uint64(gofakeit.Number(1, 10000)),
		Type:           PaymentTypeInbound,
		Amount:         gofakeit.Float64Range(100, 5000),
		Date:           now,
		State:          PaymentStateDraft,
	}
	for _, opt := range opts {
		opt(payment)
	}
	return payment
}

func TaxPeriodFixture(opts ...func(*TaxPeriod) *TaxPeriod) *TaxPeriod {
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now.AddDate(0, 1, 0)
	period := &TaxPeriod{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 100)),
		TaxYearID:      uint64(gofakeit.Number(1, 100)),
		Name:           gofakeit.MonthString() + " Period",
		DateStart:      &start,
		DateEnd:        &end,
		State:          TaxPeriodStateOpen,
		PeriodType:     TaxPeriodTypeStandard,
	}
	for _, opt := range opts {
		opt(period)
	}
	return period
}

func TaxReturnFixture(opts ...func(*TaxReturn) *TaxReturn) *TaxReturn {
	now := time.Now()
	taxReturn := &TaxReturn{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		PeriodID:       uint64(gofakeit.Number(1, 10000)),
		Type:           TaxReturnTypeSale,
		OutputTax:      gofakeit.Float64Range(100, 5000),
		InputTax:       gofakeit.Float64Range(0, 1000),
		NetPayable:     gofakeit.Float64Range(0, 5000),
		State:          TaxReturnStateDraft,
	}
	for _, opt := range opts {
		opt(taxReturn)
	}
	return taxReturn
}

func PaymentBatchFixture(opts ...func(*PaymentBatch) *PaymentBatch) *PaymentBatch {
	now := time.Now()
	batch := &PaymentBatch{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 100)),
		Name:           helper.Ptr(gofakeit.AppName()),
		JournalID:      uint64(gofakeit.Number(1, 10)),
		TotalAmount:    gofakeit.Float64Range(100, 50000),
		PaymentCount:   gofakeit.Number(1, 10),
		State:          PaymentBatchStateDraft,
		BatchDate:      &now,
	}
	for _, opt := range opts {
		opt(batch)
	}
	return batch
}

func ReconcileRuleFixture(opts ...func(*ReconcileRule) *ReconcileRule) *ReconcileRule {
	now := time.Now()
	rule := &ReconcileRule{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 100)),
		Name:           helper.Ptr(gofakeit.Word()),
		AccountID:      helper.Ptr(uint64(gofakeit.Number(1, 100))),
		MatchContact:   true,
		MatchAmount:    true,
		Active:         true,
		Sequence:       gofakeit.Number(0, 100),
	}
	for _, opt := range opts {
		opt(rule)
	}
	return rule
}

func ReminderActionFixture(opts ...func(*ReminderAction) *ReminderAction) *ReminderAction {
	now := time.Now()
	action := &ReminderAction{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ContactID: uint64(gofakeit.Number(1, 10000)),
		InvoiceID: uint64(gofakeit.Number(1, 10000)),
		LevelID:   uint64(gofakeit.Number(1, 10)),
		SentAt:    &now,
		Channel:   helper.Ptr("email"),
	}
	for _, opt := range opts {
		opt(action)
	}
	return action
}

func TaxRuleFixture(opts ...func(*TaxRule) *TaxRule) *TaxRule {
	now := time.Now()
	position := &TaxRule{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Name:           helper.Ptr(gofakeit.AppName()),
		CountryCode:    helper.Ptr(gofakeit.CountryAbr()),
		Active:         true,
	}
	for _, opt := range opts {
		opt(position)
	}
	return position
}
