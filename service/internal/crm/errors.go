package crm

import "errors"

var (
	ErrLeadNotFound            = errors.New("crm prospect not found")
	ErrLeadNameRequired        = errors.New("crm prospect name is required")
	ErrInvalidLeadType         = errors.New("crm prospect type must be prospect or opportunity")
	ErrInvalidProbability      = errors.New("crm probability must be between 0 and 100")
	ErrInvalidRevenue          = errors.New("crm expected revenue cannot be negative")
	ErrInvalidPriority         = errors.New("crm priority cannot be negative")
	ErrContactNotFound         = errors.New("crm contact does not exist")
	ErrSalespersonNotFound     = errors.New("crm salesperson does not exist")
	ErrSalesGroupNotFound      = errors.New("crm sales team does not exist")
	ErrStageNotFound           = errors.New("crm stage does not exist")
	ErrStageRequired           = errors.New("crm stage is required for an opportunity")
	ErrStageOrganization       = errors.New("crm stage does not belong to the organization")
	ErrNotLead                 = errors.New("crm record is not a prospect")
	ErrNotOpportunity          = errors.New("crm record is not an opportunity")
	ErrLeadClosed              = errors.New("crm prospect is already closed")
	ErrLostReasonRequired      = errors.New("crm lost reason is required")
	ErrNoWonStage              = errors.New("crm won stage does not exist")
	ErrActivityNotFound        = errors.New("crm activity not found")
	ErrActivitySummaryRequired = errors.New("crm activity summary is required")
	ErrActivityDone            = errors.New("crm activity is already done")
)
