package expense

import "errors"

var (
	ErrExpenseReportNotFound  = errors.New("expense report not found")
	ErrExpenseReportState     = errors.New("expense report is in an invalid state")
	ErrExpenseNoLines         = errors.New("expense report has no lines")
	ErrExpenseInvalidLine     = errors.New("expense line is invalid")
	ErrExpenseNoCategory      = errors.New("expense category not found")
	ErrExpenseNoAccount       = errors.New("expense category has no expense account")
	ErrExpenseConfig          = errors.New("expense configuration is missing")
	ErrExpenseNoBillable      = errors.New("expense report has no billable lines")
	ErrExpenseBillableProject = errors.New("expense billable lines require a single project")
	ErrExpenseNoIncomeAccount = errors.New("expense billable line has no income account")
	ErrExpenseProjectNotFound = errors.New("expense billable project not found")
)
