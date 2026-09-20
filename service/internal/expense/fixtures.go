package expense

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ExpenseReportFixture(opts ...func(*ExpenseReport) *ExpenseReport) *ExpenseReport {
	now := time.Now()
	report := &ExpenseReport{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:           gofakeit.AppName(),
		EmployeeID:     uint64(gofakeit.Number(1, 10000)),
		State:          ExpenseStateDraft,
		PaymentMode:    ExpensePaymentOwnAccount,
		TotalAmount:    gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(report)
	}
	return report
}

func ExpenseLineFixture(opts ...func(*ExpenseLine) *ExpenseLine) *ExpenseLine {
	now := time.Now()
	line := &ExpenseLine{
		Base:         model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ReportID:     uint64(gofakeit.Number(1, 10000)),
		Quantity:     gofakeit.Float64Range(1, 100),
		UnitPrice:    gofakeit.Float64Range(1, 10000),
		Amount:       gofakeit.Float64Range(1, 10000),
		Reimbursable: true,
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}
