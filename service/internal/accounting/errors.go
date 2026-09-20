package accounting

import (
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var (
	ErrNoLines          = errors.New("journal entry must have at least one line")
	ErrInvalidLine      = errors.New("journal entry line has invalid debit or credit")
	ErrUnbalanced       = errors.New("journal entry is unbalanced")
	ErrEntryNotFound    = errors.New("account entry not found")
	ErrOriginIncomplete = errors.New("origin_id requires an origin_type")

	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrAccountNotFound     = errors.New("account not found for organization")
	ErrAccountInactive     = errors.New("account is inactive")
	ErrInvoiceNoLines      = errors.New("invoice must have at least one line")
	ErrInvoiceLineQty      = errors.New("invoice line quantity must be positive")
	ErrInvoiceNotPosted    = errors.New("invoice is not posted")
	ErrInvoiceNotDraft     = errors.New("invoice is not in draft state")
	ErrInvoiceReversed     = errors.New("invoice already has a credit note")
	ErrInvoicePaid         = errors.New("invoice is already paid")
	ErrNoReceivableAccount = errors.New("no receivable account configured for organization")
	ErrNoPayableAccount    = errors.New("no payable account configured for organization")
	ErrNoRevenueAccount    = errors.New("no revenue account configured for item")
	ErrInvoiceTaxInvalid   = errors.New("invoice line references an invalid sale tax")
	ErrInvoiceSequence     = errors.New("invoice sequence is required")
	ErrBadDebtNotAllowed   = errors.New("only posted customer invoices with an open balance can be written off")

	ErrPaymentNotFound   = errors.New("payment not found")
	ErrPaymentAmount     = errors.New("payment amount must be positive")
	ErrPaymentNoInvoices = errors.New("payment must allocate at least one invoice")
	ErrOverAllocation    = errors.New("payment exceeds total open invoice balance")
	ErrNoBankAccount     = errors.New("no bank or cash account configured for journal")

	ErrPeriodNotFound      = errors.New("tax period not found")
	ErrInvalidPeriod       = errors.New("tax period date range is invalid")
	ErrInvalidPeriodState  = errors.New("tax period state is not valid")
	ErrPeriodNotClosed     = errors.New("tax period must be closed before it can be locked")
	ErrPeriodLocked        = errors.New("tax period is closed or locked")
	ErrPeriodNotOpen       = errors.New("tax period must be open to close")
	ErrPeriodAlreadyClosed = errors.New("tax period already has closing entries")

	ErrEntryNotPosted  = errors.New("account entry is not posted")
	ErrEntryNotDraft   = errors.New("account entry is not in draft state")
	ErrEntryReversed   = errors.New("account entry is already reversed")
	ErrReverserMissing = errors.New("reversal engine is not configured")
	ErrEntryImmutable  = errors.New("posted entry is immutable; use reversal")

	ErrRateNotFound            = errors.New("currency rate not found")
	ErrCurrencyMismatch        = errors.New("currency mismatch")
	ErrToleranceExceeded       = errors.New("reconcile tolerance exceeded")
	ErrReconcileWriteOff       = errors.New("reconcile write-off account required")
	ErrClosingJournalMissing   = errors.New("closing journal not configured")
	ErrInvalidRetainedEarnings = errors.New("retained earnings account is invalid")
	ErrCloseRequiresEntries    = errors.New("period close requires closing entries; use period close")

	ErrStatementNotFound  = errors.New("bank statement not found")
	ErrStatementNoLines   = errors.New("bank statement must have at least one line")
	ErrStatementCancelled = errors.New("bank statement is cancelled")
	ErrInvalidCAMT        = errors.New("invalid CAMT statement file")

	ErrLineNotFound            = errors.New("account entry line not found")
	ErrLineNotInOrganization   = errors.New("account entry line is not in the organization")
	ErrLinesDifferentAccount   = errors.New("account entry lines must be on the same account")
	ErrReconcileAmount         = errors.New("reconcile amount must be positive")
	ErrReconcileExceedsBalance = errors.New("reconcile amount exceeds the remaining line balance")

	ErrReminderLevelNotFound = errors.New("no reminder level applies to the invoice")
	ErrReminderAlreadySent   = errors.New("reminder action already sent for the invoice at this level")

	ErrBudgetNotFound     = errors.New("budget not found")
	ErrBudgetNoLines      = errors.New("budget must have at least one line")
	ErrBudgetInvalidDates = errors.New("budget date range is invalid")

	ErrTaxRuleNotFound = errors.New("fiscal position not found")

	ErrCreditMismatch = errors.New("credit note does not match the invoice")

	ErrCreditLimitExceeded = errors.New("contact credit limit exceeded")

	ErrPdcNotFound     = errors.New("pdc instrument not found")
	ErrPdcInvalidState = errors.New("pdc instrument state transition is not allowed")

	ErrWithholdingNotFound  = errors.New("withholding tax not found")
	ErrWithholdingNoAccount = errors.New("withholding tax has no account configured")
	ErrWithholdingScope     = errors.New("withholding tax scope does not match the request")

	ErrTaxReturnNotFound = errors.New("tax return not found")
	ErrTaxReturnExists   = errors.New("tax return already exists for the period")
	ErrTaxReturnNotDraft = errors.New("tax return must be in draft state")
	ErrTaxReturnNotFiled = errors.New("tax return must be filed before payment")
	ErrTaxReturnPaid     = errors.New("tax return is already paid")

	ErrScheduleNotFound    = errors.New("deferral schedule not found")
	ErrScheduleNoLines     = errors.New("deferral schedule must have at least one period")
	ErrScheduleType        = errors.New("deferral schedule type is invalid")
	ErrScheduleMethod      = errors.New("deferral schedule method is invalid")
	ErrScheduleAmount      = errors.New("deferral schedule total amount must be positive")
	ErrScheduleLineAmount  = errors.New("deferral schedule line amount must be positive")
	ErrScheduleNoAccount   = errors.New("deferral schedule must define balance sheet and pl accounts")
	ErrScheduleInvalidDate = errors.New("deferral schedule date is invalid")
	ErrScheduleNoJournal   = errors.New("no journal configured for deferral recognition")

	ErrReconcileRuleNotFound  = errors.New("reconcile rule not found")
	ErrReconcileRuleNoAccount = errors.New("reconcile rule must specify an account")

	ErrBatchNotFound   = errors.New("payment batch not found")
	ErrBatchNoPayments = errors.New("payment batch must have at least one payment")
	ErrBatchNotDraft   = errors.New("payment batch is not in draft state")

	ErrTaxYearNotFound = reference.ErrTaxYearNotFound

	ErrFxAccountNotFound = errors.New("fx gain or loss account not configured")

	ErrOrganizationNotFound = errors.New("organization not found")
	ErrBaseCurrencyMissing  = errors.New("organization base currency missing")
)
