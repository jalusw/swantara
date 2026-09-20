package reporting

import "errors"

var (
	ErrConfigMissing        = errors.New("reporting configuration missing")
	ErrUnbalancedAccrual    = errors.New("accrual lines are not balanced")
	ErrAccrualNotFound      = errors.New("accrual not found")
	ErrAccrualNotPosted     = errors.New("accrual is not posted")
	ErrAccrualReversed      = errors.New("accrual already reversed")
	ErrPeriodNotFound       = errors.New("tax period not found")
	ErrPeriodLocked         = errors.New("tax period is locked")
	ErrNoForeignPositions   = errors.New("no foreign currency positions to revalue")
	ErrBaseCurrencyMissing  = errors.New("organization base currency missing")
	ErrClosingRateMissing   = errors.New("closing rate missing")
	ErrReportingAccount     = errors.New("reporting account missing")
	ErrYearEndAlreadyClosed = errors.New("year-end retained earnings roll already posted")
)
