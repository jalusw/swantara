package asset

import "errors"

var (
	ErrAssetNotFound               = errors.New("asset not found")
	ErrAssetNameRequired           = errors.New("asset name is required")
	ErrAssetInvalidState           = errors.New("asset state transition is not allowed")
	ErrAssetCategoryNotFound       = errors.New("asset category not found")
	ErrAssetCategoryAccounts       = errors.New("asset category requires asset, depreciation, expense, gain and loss accounts")
	ErrAssetInvalidMethod          = errors.New("asset category depreciation method is invalid")
	ErrAssetInvalidPeriods         = errors.New("asset category depreciation periods must be positive")
	ErrAssetInvalidDates           = errors.New("asset dates are invalid")
	ErrAssetInvalidValues          = errors.New("asset purchase value must exceed salvage value")
	ErrAssetNotRunning             = errors.New("asset must be running to depreciate or dispose")
	ErrAssetScheduleExists         = errors.New("asset depreciation schedule already generated")
	ErrAssetNoSchedule             = errors.New("asset has no depreciation schedule")
	ErrAssetLineNotFound           = errors.New("asset depreciation line not found")
	ErrAssetNothingToPost          = errors.New("asset has no depreciation to post")
	ErrAssetLinePosted             = errors.New("asset depreciation line already posted")
	ErrAssetDisposalProceeds       = errors.New("asset disposal requires proceeds account and amount for a sale")
	ErrAssetInvoiceLineNotFound    = errors.New("supplier bill line not found")
	ErrAssetInvoiceNotSupplierBill = errors.New("asset must be registered from a posted supplier bill")
)
