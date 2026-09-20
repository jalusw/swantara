package project

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type TimesheetLookup interface {
	ListByProject(ctx context.Context, projectID uint64) ([]*payroll.Timesheet, error)
}

type ContractLookup interface {
	FindActiveByEmployee(ctx context.Context, employeeID uint64) (*payroll.EmploymentContract, error)
}

type ContactLookup interface {
	Find(ctx context.Context, id uint64) (*contacts.Contact, error)
}

type AccountLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

type InvoiceEngine interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

type InvoiceLineLookup interface {
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*accounting.InvoiceLine, error)
}
