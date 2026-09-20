package project

import "errors"

var (
	ErrProjectNotFound              = errors.New("project not found")
	ErrProjectNameRequired          = errors.New("project name is required")
	ErrProjectInvalidBillingType    = errors.New("project billing type is invalid")
	ErrProjectInvalidDates          = errors.New("project date range is invalid")
	ErrProjectInvalidState          = errors.New("project state transition is not allowed")
	ErrProjectNoBillableRate        = errors.New("project requires a billable rate for time and material billing")
	ErrProjectNotTimeMaterial       = errors.New("project billing is not time and material")
	ErrProjectNothingToBill         = errors.New("project has no unbilled billable timesheets")
	ErrProjectNoRevenueAccount      = errors.New("project requires an income account for the organization")
	ErrProjectContactNotFound       = errors.New("project contact not found")
	ErrProjectDimensionNotFound     = errors.New("project dimension account not found")
	ErrProjectCostNotAttributable   = errors.New("project cost is not attributable for an employee without an active contract")
	ErrProjectTaskNotFound          = errors.New("project task not found")
	ErrProjectTaskNameRequired      = errors.New("project task name is required")
	ErrProjectInvalidHours          = errors.New("project task planned hours must not be negative")
	ErrProjectMilestoneNotFound     = errors.New("project milestone not found")
	ErrProjectMilestoneNameRequired = errors.New("project milestone name is required")
)
