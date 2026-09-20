package project

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ProjectService struct {
	projects      ProjectDAO
	tasks         ProjectTaskDAO
	milestones    ProjectMilestoneDAO
	invoiceLines  ProjectInvoiceLineDAO
	timesheets    TimesheetLookup
	contracts     ContractLookup
	contacts      ContactLookup
	accounts      AccountLookup
	dimensions    dao.CRUD[reference.Dimension]
	invoices      InvoiceEngine
	invoiceLookup InvoiceLineLookup
	tx            db.Transactioner
	machine       state.Machine
	now           func() time.Time
}

func NewProjectService(
	projects ProjectDAO,
	tasks ProjectTaskDAO,
	milestones ProjectMilestoneDAO,
	invoiceLines ProjectInvoiceLineDAO,
	timesheets TimesheetLookup,
	contracts ContractLookup,
	contacts ContactLookup,
	accounts AccountLookup,
	dimensions dao.CRUD[reference.Dimension],
	invoices InvoiceEngine,
	invoiceLookup InvoiceLineLookup,
	tx db.Transactioner,
) ProjectService {
	return ProjectService{
		projects:      projects,
		tasks:         tasks,
		milestones:    milestones,
		invoiceLines:  invoiceLines,
		timesheets:    timesheets,
		contracts:     contracts,
		contacts:      contacts,
		accounts:      accounts,
		dimensions:    dimensions,
		invoices:      invoices,
		invoiceLookup: invoiceLookup,
		tx:            tx,
		machine: state.NewMachine(
			state.Transition{From: model.Status(ProjectStateDraft), To: model.Status(ProjectStateOpen)},
			state.Transition{From: model.Status(ProjectStateDraft), To: model.Status(ProjectStateCancelled)},
			state.Transition{From: model.Status(ProjectStateOpen), To: model.Status(ProjectStateClosed)},
			state.Transition{From: model.Status(ProjectStateOpen), To: model.Status(ProjectStateCancelled)},
		),
		now: time.Now,
	}
}

func (s ProjectService) validateProject(ctx context.Context, project *Project) error {
	if project.Name == "" {
		return ErrProjectNameRequired
	}
	switch project.BillingType {
	case BillingTypeFixed, BillingTypeTimeMaterial, BillingTypeMilestone:
	default:
		return ErrProjectInvalidBillingType
	}
	if project.DateStart != nil && project.DateEnd != nil && project.DateEnd.Before(*project.DateStart) {
		return ErrProjectInvalidDates
	}
	contact, err := s.contacts.Find(ctx, project.ContactID)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrProjectContactNotFound
	}
	if project.DimensionID != nil {
		account, err := s.dimensions.Find(ctx, *project.DimensionID)
		if err != nil {
			return err
		}
		if account == nil {
			return ErrProjectDimensionNotFound
		}
	}
	return nil
}

func (s ProjectService) List(ctx context.Context, q *query.Query) (*query.Page[Project], error) {
	return s.projects.List(ctx, q)
}

func (s ProjectService) Find(ctx context.Context, id uint64) (*Project, error) {
	return s.projects.Find(ctx, id)
}

func (s ProjectService) ListTasksByProject(ctx context.Context, projectID uint64) ([]*ProjectTask, error) {
	return s.tasks.ListByProject(ctx, projectID)
}

func (s ProjectService) ListMilestonesByProject(ctx context.Context, projectID uint64) ([]*ProjectMilestone, error) {
	return s.milestones.ListByProject(ctx, projectID)
}

func (s ProjectService) CreateProject(ctx context.Context, project *Project) (*Project, error) {
	if err := s.validateProject(ctx, project); err != nil {
		return nil, err
	}
	project.State = ProjectStateDraft
	return s.projects.Create(ctx, project)
}

func (s ProjectService) UpdateProject(ctx context.Context, project *Project) (*Project, error) {
	existing, err := s.projects.Find(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrProjectNotFound
	}
	if err := s.validateProject(ctx, project); err != nil {
		return nil, err
	}
	project.State = existing.State
	return s.projects.Update(ctx, project)
}

func (s ProjectService) SetProjectState(ctx context.Context, projectID uint64, next string) (*Project, error) {
	project, err := s.projects.Find(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	if err := s.machine.TryTransition(model.Status(project.State), model.Status(next)); err != nil {
		return nil, ErrProjectInvalidState
	}
	project.State = next
	return s.projects.Update(ctx, project)
}

func (s ProjectService) CreateTask(ctx context.Context, task *ProjectTask) (*ProjectTask, error) {
	project, err := s.projects.Find(ctx, task.ProjectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	if project.State != ProjectStateOpen {
		return nil, ErrProjectInvalidState
	}
	if task.Name == "" {
		return nil, ErrProjectTaskNameRequired
	}
	if task.PlannedHours < 0 {
		return nil, ErrProjectInvalidHours
	}
	if task.Stage == "" {
		task.Stage = TaskStageBacklog
	}
	return s.tasks.Create(ctx, task)
}

func (s ProjectService) UpdateTask(ctx context.Context, task *ProjectTask) (*ProjectTask, error) {
	existing, err := s.tasks.Find(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrProjectTaskNotFound
	}
	if task.Name == "" {
		return nil, ErrProjectTaskNameRequired
	}
	if task.PlannedHours < 0 {
		return nil, ErrProjectInvalidHours
	}
	if task.Stage == "" {
		task.Stage = TaskStageBacklog
	}
	return s.tasks.Update(ctx, task)
}

func (s ProjectService) CreateMilestone(ctx context.Context, milestone *ProjectMilestone) (*ProjectMilestone, error) {
	project, err := s.projects.Find(ctx, milestone.ProjectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	if milestone.Name == "" {
		return nil, ErrProjectMilestoneNameRequired
	}
	return s.milestones.Create(ctx, milestone)
}

func (s ProjectService) SetMilestoneReached(ctx context.Context, milestoneID uint64, reached bool) (*ProjectMilestone, error) {
	milestone, err := s.milestones.Find(ctx, milestoneID)
	if err != nil {
		return nil, err
	}
	if milestone == nil {
		return nil, ErrProjectMilestoneNotFound
	}
	milestone.Reached = reached
	return s.milestones.Update(ctx, milestone)
}

type ProjectSummary struct {
	ProjectID      uint64
	PlannedHours   float64
	EffectiveHours float64
	Utilization    float64
	BilledAmount   float64
	UnbilledAmount float64
	BillableAmount float64
	CostAmount     float64
	MarginAmount   float64
}

func (s ProjectService) Summary(ctx context.Context, projectID uint64) (ProjectSummary, error) {
	project, err := s.projects.Find(ctx, projectID)
	if err != nil {
		return ProjectSummary{}, err
	}
	if project == nil {
		return ProjectSummary{}, ErrProjectNotFound
	}

	tasks, err := s.tasks.ListByProject(ctx, projectID)
	if err != nil {
		return ProjectSummary{}, err
	}
	timesheets, err := s.timesheets.ListByProject(ctx, projectID)
	if err != nil {
		return ProjectSummary{}, err
	}
	invoiceLines, err := s.invoiceLines.ListByProject(ctx, projectID)
	if err != nil {
		return ProjectSummary{}, err
	}

	var planned, effective float64
	for _, task := range tasks {
		planned += task.PlannedHours
	}
	for _, timesheet := range timesheets {
		effective += timesheet.Hours
	}

	billed := 0.0
	for _, line := range invoiceLines {
		billed += line.Amount
	}
	unbilled := 0.0
	for _, timesheet := range timesheets {
		if s.isTimesheetBilled(timesheet, invoiceLines) {
			continue
		}
		unbilled += timesheet.Hours * project.BillableRate
	}
	billable := billed + unbilled

	cost, err := s.projectCost(ctx, timesheets)
	if err != nil {
		return ProjectSummary{}, err
	}

	summary := ProjectSummary{
		ProjectID:      projectID,
		PlannedHours:   planned,
		EffectiveHours: effective,
		BilledAmount:   billed,
		UnbilledAmount: unbilled,
		BillableAmount: billable,
		CostAmount:     cost,
	}
	if planned > 0 {
		summary.Utilization = effective / planned
	}
	summary.MarginAmount = billable - cost
	return summary, nil
}

func (s ProjectService) isTimesheetBilled(timesheet *payroll.Timesheet, invoiceLines []*ProjectInvoiceLine) bool {
	for _, line := range invoiceLines {
		if line.TimesheetID == timesheet.ID {
			return true
		}
	}
	return false
}

func (s ProjectService) projectCost(ctx context.Context, timesheets []*payroll.Timesheet) (float64, error) {
	total := 0.0
	for _, timesheet := range timesheets {
		contract, err := s.contracts.FindActiveByEmployee(ctx, timesheet.EmployeeID)
		if err != nil {
			return 0, err
		}
		if contract == nil {
			return 0, ErrProjectCostNotAttributable
		}
		total += timesheet.Hours * (contract.Wage / StandardMonthlyHours)
	}
	return total, nil
}

func (s ProjectService) revenueAccount(ctx context.Context, organizationID uint64) (uint64, error) {
	accounts, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: "income"}}})
	if err != nil {
		return 0, err
	}
	for _, account := range accounts.Items {
		if account.OrganizationID == organizationID && account.Active {
			return account.ID, nil
		}
	}
	return 0, ErrProjectNoRevenueAccount
}

type BillTimeMaterialRequest struct {
	OrganizationID uint64
	ProjectID      uint64
	JournalID      uint64
	Date           time.Time
	TaxIDs         helper.Int64Array
}

func (s ProjectService) BillTimeMaterial(ctx context.Context, request BillTimeMaterialRequest) (*accounting.Invoice, error) {
	project, err := s.projects.Find(ctx, request.ProjectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	if project.State != ProjectStateOpen {
		return nil, ErrProjectInvalidState
	}
	if project.BillingType != BillingTypeTimeMaterial {
		return nil, ErrProjectNotTimeMaterial
	}
	if project.BillableRate <= 0 {
		return nil, ErrProjectNoBillableRate
	}

	timesheets, err := s.timesheets.ListByProject(ctx, request.ProjectID)
	if err != nil {
		return nil, err
	}
	invoiceLines, err := s.invoiceLines.ListByProject(ctx, request.ProjectID)
	if err != nil {
		return nil, err
	}
	unbilled := make([]*payroll.Timesheet, 0, len(timesheets))
	for _, timesheet := range timesheets {
		if !s.isTimesheetBilled(timesheet, invoiceLines) {
			unbilled = append(unbilled, timesheet)
		}
	}
	if len(unbilled) == 0 {
		return nil, ErrProjectNothingToBill
	}

	revenueAccount, err := s.revenueAccount(ctx, request.OrganizationID)
	if err != nil {
		return nil, err
	}

	lines := make([]accounting.InvoiceLineRequest, len(unbilled))
	for i, timesheet := range unbilled {
		lines[i] = accounting.InvoiceLineRequest{
			Description: project.Name,
			Qty:         timesheet.Hours,
			UnitPrice:   project.BillableRate,
			TaxIDs:      request.TaxIDs,
			AccountID:   revenueAccount,
			DimensionID: project.DimensionID,
		}
	}

	invoice, err := s.invoices.Create(ctx, accounting.CreateInvoiceRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		ContactID:      project.ContactID,
		Date:           request.Date,
		Reference:      project.Name,
		Lines:          lines,
	})
	if err != nil {
		return nil, err
	}

	createdLines, err := s.invoiceLookup.ListByInvoice(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	for i, timesheet := range unbilled {
		amount := timesheet.Hours * project.BillableRate
		join := &ProjectInvoiceLine{
			ProjectID:   project.ID,
			InvoiceID:   invoice.ID,
			TimesheetID: timesheet.ID,
			Qty:         timesheet.Hours,
			UnitPrice:   project.BillableRate,
			Amount:      amount,
		}
		if i < len(createdLines) {
			join.InvoiceLineID = helper.Ptr(createdLines[i].ID)
		}
		if _, err := s.invoiceLines.Create(ctx, join); err != nil {
			return nil, err
		}
	}
	return invoice, nil
}
