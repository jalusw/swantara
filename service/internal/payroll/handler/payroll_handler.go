package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type SalaryRuleHandler struct {
	svc payroll.SalaryRuleService
}

func NewSalaryRuleHandler(svc payroll.SalaryRuleService) SalaryRuleHandler {
	return SalaryRuleHandler{svc: svc}
}

type SalaryRuleResponse struct {
	ID              uint64   `json:"id"`
	OrganizationID  *uint64  `json:"organization_id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Category        *string  `json:"category"`
	ComputeType     *string  `json:"compute_type"`
	Amount          *float64 `json:"amount"`
	Formula         *string  `json:"formula"`
	AccountDebitID  *uint64  `json:"account_debit_id"`
	AccountCreditID *uint64  `json:"account_credit_id"`
}

func newSalaryRuleResponse(rule *reference.SalaryRule) SalaryRuleResponse {
	return SalaryRuleResponse{
		ID:              rule.ID,
		OrganizationID:  rule.OrganizationID,
		Code:            rule.Code,
		Name:            rule.Name,
		Category:        rule.Category,
		ComputeType:     rule.ComputeType,
		Amount:          rule.Amount,
		Formula:         rule.Formula,
		AccountDebitID:  rule.AccountDebitID,
		AccountCreditID: rule.AccountCreditID,
	}
}

type CreateSalaryRuleRequest struct {
	OrganizationID  *uint64  `json:"organization_id"`
	Code            string   `json:"code" validate:"required"`
	Name            string   `json:"name" validate:"required"`
	Category        string   `json:"category" validate:"required"`
	ComputeType     string   `json:"compute_type" validate:"required"`
	Amount          *float64 `json:"amount"`
	Formula         *string  `json:"formula"`
	AccountDebitID  *uint64  `json:"account_debit_id" validate:"required,gt=0"`
	AccountCreditID *uint64  `json:"account_credit_id" validate:"required,gt=0"`
}

type ListSalaryRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []SalaryRuleResponse `json:"data"`
}

type SalaryRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SalaryRuleResponse `json:"data"`
}

var salaryRuleQueryAllowlist = map[string]struct{}{
	"code":            {},
	"category":        {},
	"compute_type":    {},
	"organization_id": {},
}

// @Summary List salary rules
// @Description Lists salary rules with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to allowlisted fields (code, category, and compute type), and the organization filter is forced to the caller's tenant so cross-tenant salary rules are never returned.
// @Tags Salary Rules
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListSalaryRulesResponseEnvelope "Salary rules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/salary-rules [get]
func (h SalaryRuleHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, salaryRuleQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("salary rule list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve salary rules.", err)
	}
	items := make([]SalaryRuleResponse, len(page.Items))
	for i, rule := range page.Items {
		items[i] = newSalaryRuleResponse(rule)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Salary rules retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get salary rule
// @Description Gets a single salary rule by its id. The salary rule must belong to the caller's organization, and a 404 not found is returned when it does not, even if a salary rule with that id exists in another tenant.
// @Tags Salary Rules
// @Accept json
// @Produce json
// @Param id path integer true "Salary rule ID"
// @Success 200 {object} SalaryRuleResponseEnvelope "Salary rule retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Salary rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/salary-rules/{id} [get]
func (h SalaryRuleHandler) Get(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid salary rule id provided.", nil)
	}
	rule, err := h.svc.Find(c, ruleID)
	if err != nil {
		httpx.RequestLog(c).Error("salary rule lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get salary rule.", err)
	}
	if rule == nil || !httpx.OwnsTenant(c, rule.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Salary rule not found.")
	}
	return httpx.CreateSuccessResponse(c, "Salary rule retrieved successfully.", newSalaryRuleResponse(rule))
}

// @Summary Create salary rule
// @Description Creates a new salary rule under the caller's organization defining a code, category (earning or deduction), computation type (fixed, percent, or formula), and the debit and credit accounts used when posting payroll. The code, category, and compute type must be valid, and both referenced accounts must exist.
// @Tags Salary Rules
// @Accept json
// @Produce json
// @Param request body CreateSalaryRuleRequest true "Salary rule details"
// @Success 201 {object} SalaryRuleResponseEnvelope "Salary rule created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/salary-rules [post]
func (h SalaryRuleHandler) Create(c fiber.Ctx) error {
	var request CreateSalaryRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	rule, err := h.svc.Create(c, &reference.SalaryRule{
		OrganizationID:  organizationID,
		Code:            request.Code,
		Name:            request.Name,
		Category:        helper.Ptr(request.Category),
		ComputeType:     helper.Ptr(request.ComputeType),
		Amount:          request.Amount,
		Formula:         request.Formula,
		AccountDebitID:  request.AccountDebitID,
		AccountCreditID: request.AccountCreditID,
	})
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Salary rule created successfully.", newSalaryRuleResponse(rule))
}

type UpdateSalaryRuleRequest struct {
	Code            string   `json:"code" validate:"required"`
	Name            string   `json:"name" validate:"required"`
	Category        string   `json:"category" validate:"required"`
	ComputeType     string   `json:"compute_type" validate:"required"`
	Amount          *float64 `json:"amount"`
	Formula         *string  `json:"formula"`
	AccountDebitID  *uint64  `json:"account_debit_id" validate:"required,gt=0"`
	AccountCreditID *uint64  `json:"account_credit_id" validate:"required,gt=0"`
}

// @Summary Update salary rule
// @Description Updates an existing salary rule's code, category, computation settings, and debit/credit accounts. The salary rule must belong to the caller's organization (404 otherwise), and the same validations applied on creation are enforced on the updated payload.
// @Tags Salary Rules
// @Accept json
// @Produce json
// @Param id path integer true "Salary rule ID"
// @Param request body UpdateSalaryRuleRequest true "Salary rule details"
// @Success 200 {object} SalaryRuleResponseEnvelope "Salary rule updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Salary rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/salary-rules/{id} [put]
func (h SalaryRuleHandler) Update(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid salary rule id provided.", nil)
	}
	rule, err := h.svc.Find(c, ruleID)
	if err != nil {
		httpx.RequestLog(c).Error("salary rule lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get salary rule.", err)
	}
	if rule == nil || !httpx.OwnsTenant(c, rule.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Salary rule not found.")
	}
	var request UpdateSalaryRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	rule.Code = request.Code
	rule.Name = request.Name
	rule.Category = helper.Ptr(request.Category)
	rule.ComputeType = helper.Ptr(request.ComputeType)
	rule.Amount = request.Amount
	rule.Formula = request.Formula
	rule.AccountDebitID = request.AccountDebitID
	rule.AccountCreditID = request.AccountCreditID
	updated, err := h.svc.Update(c, rule)
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Salary rule updated successfully.", newSalaryRuleResponse(updated))
}

// @Summary Delete salary rule
// @Description Deletes a salary rule by id. The salary rule must exist within the caller's organization or a 404 response is returned, and a successful deletion returns no content.
// @Tags Salary Rules
// @Accept json
// @Produce json
// @Param id path integer true "Salary rule ID"
// @Success 204 {object} httpx.EmptyEnvelope "Salary rule deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Salary rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/salary-rules/{id} [delete]
func (h SalaryRuleHandler) Delete(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid salary rule id provided.", nil)
	}
	rule, err := h.svc.Find(c, ruleID)
	if err != nil {
		httpx.RequestLog(c).Error("salary rule lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get salary rule.", err)
	}
	if rule == nil || !httpx.OwnsTenant(c, rule.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Salary rule not found.")
	}
	if err := h.svc.Delete(c, ruleID); err != nil {
		httpx.RequestLog(c).Error("salary rule delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete salary rule.", err)
	}
	return httpx.CreateNoContentResponse(c)
}

type PayrollRunHandler struct {
	svc payroll.PayrollService
}

func NewPayrollRunHandler(svc payroll.PayrollService) PayrollRunHandler {
	return PayrollRunHandler{svc: svc}
}

type PayrollRunResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID uint64  `json:"organization_id"`
	Name           *string `json:"name"`
	PeriodStart    string  `json:"period_start"`
	PeriodEnd      string  `json:"period_end"`
	State          string  `json:"state"`
}

func newPayrollRunResponse(run *payroll.PayrollRun) PayrollRunResponse {
	return PayrollRunResponse{
		ID:             run.ID,
		OrganizationID: run.OrganizationID,
		Name:           run.Name,
		PeriodStart:    run.PeriodStart.Format("2006-01-02"),
		PeriodEnd:      run.PeriodEnd.Format("2006-01-02"),
		State:          run.State,
	}
}

type CreatePayrollRunRequest struct {
	OrganizationID uint64 `json:"organization_id" validate:"required,gt=0"`
	PeriodStart    string `json:"period_start" validate:"required"`
	PeriodEnd      string `json:"period_end" validate:"required"`
}

type ListPayrollRunsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []PayrollRunResponse `json:"data"`
}

type PayrollRunResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PayrollRunResponse `json:"data"`
}

type PayrollRunDetailResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PayrollRunDetailResponse `json:"data"`
}

var payrollRunQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"state":           {},
	"period_start":    {},
	"period_end":      {},
}

// @Summary List payroll runs
// @Description Lists payroll runs with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to an allowlisted set of fields, and the organization filter is forced to the caller's tenant so cross-tenant payroll runs are never returned.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPayrollRunsResponseEnvelope "Payroll runs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs [get]
func (h PayrollRunHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, payrollRunQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListRuns(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payroll run list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payroll runs.", err)
	}
	items := make([]PayrollRunResponse, len(page.Items))
	for i, run := range page.Items {
		items[i] = newPayrollRunResponse(run)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Payroll runs retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get payroll run
// @Description Gets a single payroll run by its id along with its payslips and their line items. The payroll run must belong to the caller's organization (404 otherwise), and payslips are only loaded for the retrieved run.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param id path integer true "Payroll run ID"
// @Success 200 {object} PayrollRunDetailResponseEnvelope "Payroll run retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payroll run not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs/{id} [get]
func (h PayrollRunHandler) Get(c fiber.Ctx) error {
	runID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payroll run id provided.", nil)
	}
	run, err := h.svc.FindRun(c, runID)
	if err != nil {
		httpx.RequestLog(c).Error("payroll run lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payroll run.", err)
	}
	if run == nil || !httpx.OwnsTenant(c, &run.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Payroll run not found.")
	}
	payslips, err := h.svc.ListPayslipsByRun(c, runID)
	if err != nil {
		httpx.RequestLog(c).Error("payslip lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payslips.", err)
	}
	payslipItems := make([]PayslipResponse, len(payslips))
	for i, payslip := range payslips {
		lines, err := h.svc.ListPayslipLines(c, payslip.ID)
		if err != nil {
			httpx.RequestLog(c).Error("payslip line lookup failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to get payslip lines.", err)
		}
		payslipItems[i] = newPayslipResponse(payslip, lines)
	}
	return httpx.CreateSuccessResponse(c, "Payroll run retrieved successfully.", PayrollRunDetailResponse{
		Run:      newPayrollRunResponse(run),
		Payslips: payslipItems,
	})
}

type PayrollRunDetailResponse struct {
	Run      PayrollRunResponse `json:"run"`
	Payslips []PayslipResponse  `json:"payslips"`
}

// @Summary Create payroll run
// @Description Creates a new draft payroll run for a pay period, computing and storing payslips for every active employee with an active employment contract. Period dates must be in YYYY-MM-DD format with period_start on or before period_end, and a 422 response is returned when no eligible employees are found.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param request body CreatePayrollRunRequest true "Payroll run details"
// @Success 201 {object} PayrollRunResponseEnvelope "Payroll run created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs [post]
func (h PayrollRunHandler) Create(c fiber.Ctx) error {
	var request CreatePayrollRunRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	periodStart, err := helper.ParseDate(&request.PeriodStart)
	if err != nil || periodStart == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Period start must be in YYYY-MM-DD format.", nil)
	}
	periodEnd, err := helper.ParseDate(&request.PeriodEnd)
	if err != nil || periodEnd == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Period end must be in YYYY-MM-DD format.", nil)
	}
	run, err := h.svc.CreateRun(c, payroll.CreateRunRequest{
		OrganizationID: request.OrganizationID,
		PeriodStart:    *periodStart,
		PeriodEnd:      *periodEnd,
	})
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Payroll run created successfully.", newPayrollRunResponse(run))
}

type ConfirmPayrollRunRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
}

// @Summary Confirm payroll run
// @Description Confirms a draft payroll run, posting the payroll accrual journal entries to each salary rule's debit and credit accounts across all payslips. The run must be in draft state, the journal must exist, and the date must be in YYYY-MM-DD format; on success the run movements to confirmed and its payslips to posted.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param id path integer true "Payroll run ID"
// @Param request body ConfirmPayrollRunRequest true "Journal and date details"
// @Success 200 {object} PayrollRunResponseEnvelope "Payroll run confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Payroll cannot be processed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs/{id}/confirm [post]
func (h PayrollRunHandler) Confirm(c fiber.Ctx) error {
	runID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payroll run id provided.", nil)
	}
	var request ConfirmPayrollRunRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(&request.Date)
	if err != nil || date == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	run, err := h.svc.Confirm(c, runID, request.JournalID, *date)
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Payroll run confirmed successfully.", newPayrollRunResponse(run))
}

type PayPayrollRunRequest struct {
	JournalID           uint64 `json:"journal_id" validate:"required,gt=0"`
	NetPayableAccountID uint64 `json:"net_payable_account_id" validate:"required,gt=0"`
	Date                string `json:"date" validate:"required"`
}

// @Summary Pay payroll run
// @Description Marks a confirmed payroll run as paid, posting a payment journal entry that debits the given net payable account and credits the bank account linked to the journal. The run must be in confirmed state, the journal must define a bank account, and the combined net pay of all payslips must be positive; on success the run movements to paid.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param id path integer true "Payroll run ID"
// @Param request body PayPayrollRunRequest true "Journal, net payable account and date details"
// @Success 200 {object} PayrollRunResponseEnvelope "Payroll run paid successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Payroll cannot be processed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs/{id}/pay [post]
func (h PayrollRunHandler) Pay(c fiber.Ctx) error {
	runID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payroll run id provided.", nil)
	}
	var request PayPayrollRunRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(&request.Date)
	if err != nil || date == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	run, err := h.svc.Pay(c, runID, request.JournalID, request.NetPayableAccountID, *date)
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Payroll run paid successfully.", newPayrollRunResponse(run))
}

// @Summary Close payroll run
// @Description Closes a paid payroll run, finalizing the pay period and moving the run to closed state. Only runs in paid state can be closed, and attempting to close a run in any other state returns a 422 response.
// @Tags Payroll Runs
// @Accept json
// @Produce json
// @Param id path integer true "Payroll run ID"
// @Success 200 {object} PayrollRunResponseEnvelope "Payroll run closed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Payroll cannot be processed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payroll-runs/{id}/close [post]
func (h PayrollRunHandler) Close(c fiber.Ctx) error {
	runID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payroll run id provided.", nil)
	}
	run, err := h.svc.Close(c, runID)
	if err != nil {
		return writePayrollError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Payroll run closed successfully.", newPayrollRunResponse(run))
}

type PayslipHandler struct {
	svc payroll.PayrollService
}

func NewPayslipHandler(svc payroll.PayrollService) PayslipHandler {
	return PayslipHandler{svc: svc}
}

type PayslipLineResponse struct {
	ID        uint64  `json:"id"`
	PayslipID uint64  `json:"payslip_id"`
	RuleID    uint64  `json:"rule_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
}

func newPayslipLineResponse(line *payroll.PayslipLine) PayslipLineResponse {
	return PayslipLineResponse{
		ID:        line.ID,
		PayslipID: line.PayslipID,
		RuleID:    line.RuleID,
		Code:      line.Code,
		Name:      line.Name,
		Category:  line.Category,
		Amount:    line.Amount,
	}
}

type PayslipResponse struct {
	ID         uint64                `json:"id"`
	RunID      uint64                `json:"run_id"`
	EmployeeID uint64                `json:"employee_id"`
	ContractID uint64                `json:"contract_id"`
	Gross      float64               `json:"gross"`
	Net        float64               `json:"net"`
	EntryID    *uint64               `json:"entry_id"`
	State      string                `json:"state"`
	Lines      []PayslipLineResponse `json:"lines"`
}

func newPayslipResponse(payslip *payroll.Payslip, lines []*payroll.PayslipLine) PayslipResponse {
	lineItems := make([]PayslipLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newPayslipLineResponse(line)
	}
	return PayslipResponse{
		ID:         payslip.ID,
		RunID:      payslip.RunID,
		EmployeeID: payslip.EmployeeID,
		ContractID: payslip.ContractID,
		Gross:      payslip.Gross,
		Net:        payslip.Net,
		EntryID:    payslip.EntryID,
		State:      payslip.State,
		Lines:      lineItems,
	}
}

type ListPayslipsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []PayslipResponse `json:"data"`
}

type PayslipResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PayslipResponse `json:"data"`
}

var payslipQueryAllowlist = map[string]struct{}{
	"run_id":      {},
	"employee_id": {},
	"state":       {},
}

// @Summary List payslips
// @Description Lists payslips with pagination, sorting, and filtering, including each payslip's line items. Queries are restricted to allowlisted fields (run, employee, and state), and invalid query parameters return a 422 response.
// @Tags Payslips
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPayslipsResponseEnvelope "Payslips retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payslips [get]
func (h PayslipHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, payslipQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	page, err := h.svc.ListPayslips(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payslip list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payslips.", err)
	}
	items := make([]PayslipResponse, len(page.Items))
	for i, payslip := range page.Items {
		lines, err := h.svc.ListPayslipLines(c, payslip.ID)
		if err != nil {
			httpx.RequestLog(c).Error("payslip line lookup failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to get payslip lines.", err)
		}
		items[i] = newPayslipResponse(payslip, lines)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Payslips retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get payslip
// @Description Gets a single payslip by its id along with its line items. The id must be a valid positive integer, and a 404 response is returned when no matching payslip exists.
// @Tags Payslips
// @Accept json
// @Produce json
// @Param id path integer true "Payslip ID"
// @Success 200 {object} PayslipResponseEnvelope "Payslip retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payslip not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payslips/{id} [get]
func (h PayslipHandler) Get(c fiber.Ctx) error {
	payslipID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payslip id provided.", nil)
	}
	payslip, err := h.svc.FindPayslip(c, payslipID)
	if err != nil {
		httpx.RequestLog(c).Error("payslip lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payslip.", err)
	}
	if payslip == nil {
		return httpx.CreateNotFoundResponse(c, "Payslip not found.")
	}
	lines, err := h.svc.ListPayslipLines(c, payslip.ID)
	if err != nil {
		httpx.RequestLog(c).Error("payslip line lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payslip lines.", err)
	}
	return httpx.CreateSuccessResponse(c, "Payslip retrieved successfully.", newPayslipResponse(payslip, lines))
}

func writePayrollError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, payroll.ErrRunNotFound), errors.Is(err, payroll.ErrPayslipNotFound),
		errors.Is(err, payroll.ErrRuleNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, payroll.ErrRunState), errors.Is(err, payroll.ErrRunPeriod),
		errors.Is(err, payroll.ErrRunOrganization), errors.Is(err, payroll.ErrRunNoPayslips),
		errors.Is(err, payroll.ErrRuleCode), errors.Is(err, payroll.ErrRuleCategory),
		errors.Is(err, payroll.ErrRuleComputeType), errors.Is(err, payroll.ErrRuleAccounts),
		errors.Is(err, payroll.ErrRuleAccount), errors.Is(err, payroll.ErrNoBankAccount),
		errors.Is(err, payroll.ErrNoNetPayable):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payroll cannot be processed.", nil)
	case errors.Is(err, payroll.ErrRunSequence):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payroll run sequence is not configured.", nil)
	default:
		httpx.RequestLog(c).Error("payroll write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process payroll.", err)
	}
}
