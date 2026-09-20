import type { AxiosInstance } from "axios";
import { endpoints } from "./endpoints";
import type {
  Account,
  AccountingStandard,
  Accrual,
  AccrualResult,
  AgingRow,
  ApprovalRequest,
  ArApKpi,
  Attachment,
  Attendance,
  BalanceSheet,
  BillTimeMaterialRequest,
  Carrier,
  CashFlow,
  CashKpi,
  CheckInRequest,
  CheckOutRequest,
  ClosePeriodRequest,
  CloseResult,
  ConfirmPayrollRunRequest,
  ConvertUnitRequest,
  CreateAccountRequest,
  CreateAccrualRequest,
  CreateApprovalRequestRequest,
  CreateCarrierRequest,
  CreateContractRequest,
  CreateDepartmentRequest,
  CreateDimensionRequest,
  CreateEmployeeRequest,
  CreateFxRateRequest,
  CreateJobPositionRequest,
  CreateJournalRequest,
  CreateLeaveRequestRequest,
  CreateLeaveTypeRequest,
  CreateMessageRequest,
  CreatePaymentTermRequest,
  CreatePayrollRunRequest,
  CreateProjectMilestoneRequest,
  CreateProjectRequest,
  CreateProjectTaskRequest,
  CreateSalaryRuleRequest,
  CreateTaxRequest,
  CreateTaxYearRequest,
  CreateTimesheetRequest,
  CreateUnitCategoryRequest,
  CreateUnitRequest,
  Currency,
  DecideApprovalRequestRequest,
  Department,
  Dimension,
  Employee,
  EmploymentContract,
  FinanceKpi,
  FxRate,
  FxRevaluation,
  InventoryKpi,
  InventoryRatioKpi,
  InventoryValueRow,
  JobPosition,
  Journal,
  LeaveBalance,
  LeaveRequest,
  LeaveType,
  ListQuery,
  ManufacturingKpi,
  Message,
  PaymentSplit,
  PaymentTerm,
  PayPayrollRunRequest,
  PayrollKpi,
  PayrollRun,
  Payslip,
  PipelineKpi,
  ProcurementKpi,
  ProfitAndLoss,
  Project,
  ProjectKpi,
  ProjectMilestone,
  ProjectSummary,
  ProjectTask,
  ResolveFxRateRequest,
  ResolveFxRateResponse,
  RevaluationResult,
  RevalueRequest,
  ReverseAccrualsRequest,
  SalaryRule,
  SalesKpi,
  SetMilestoneReachedRequest,
  SetProjectStateRequest,
  SplitPaymentTermRequest,
  SubscriptionKpi,
  SuccessEnvelope,
  Tax,
  TaxYear,
  Timesheet,
  TrialBalance,
  Unit,
  UnitCategory,
  UpdateAccountRequest,
  UpdateCarrierRequest,
  UpdateContractRequest,
  UpdateDimensionRequest,
  UpdateEmployeeRequest,
  UpdateFxRateRequest,
  UpdatePaymentTermRequest,
  UpdateProjectRequest,
  UpdateProjectTaskRequest,
  UpdateSalaryRuleRequest,
  UpdateTaxRequest,
  UpdateTaxYearRequest,
  UpdateUnitRequest,
  UploadAttachmentRequest,
  YearEndRollRequest,
  YearEndRollResult,
} from "./types";
import { withListMeta } from "./types";

export class Departments {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ departments: Department[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ departments: Department[] }>>(
      endpoints.departments.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ department: Department }> {
    const response = await this.axios.get<SuccessEnvelope<{ department: Department }>>(
      endpoints.departments.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateDepartmentRequest,
  ): Promise<{ department: Department }> {
    const response = await this.axios.post<SuccessEnvelope<{ department: Department }>>(
      endpoints.departments.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: Partial<CreateDepartmentRequest>,
  ): Promise<{ department: Department }> {
    const response = await this.axios.put<SuccessEnvelope<{ department: Department }>>(
      endpoints.departments.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.departments.delete(String(organizationId), String(id)));
  }
}

export class JobPositions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ jobPositions: JobPosition[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ jobPositions: JobPosition[] }>>(
      endpoints.jobPositions.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ jobPosition: JobPosition }> {
    const response = await this.axios.get<SuccessEnvelope<{ jobPosition: JobPosition }>>(
      endpoints.jobPositions.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateJobPositionRequest,
  ): Promise<{ jobPosition: JobPosition }> {
    const response = await this.axios.post<SuccessEnvelope<{ jobPosition: JobPosition }>>(
      endpoints.jobPositions.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: Partial<CreateJobPositionRequest>,
  ): Promise<{ jobPosition: JobPosition }> {
    const response = await this.axios.put<SuccessEnvelope<{ jobPosition: JobPosition }>>(
      endpoints.jobPositions.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.jobPositions.delete(String(organizationId), String(id)));
  }
}

export class LeaveTypes {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ leaveTypes: LeaveType[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ leaveTypes: LeaveType[] }>>(
      endpoints.leaveTypes.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ leaveType: LeaveType }> {
    const response = await this.axios.get<SuccessEnvelope<{ leaveType: LeaveType }>>(
      endpoints.leaveTypes.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateLeaveTypeRequest,
  ): Promise<{ leaveType: LeaveType }> {
    const response = await this.axios.post<SuccessEnvelope<{ leaveType: LeaveType }>>(
      endpoints.leaveTypes.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: Partial<CreateLeaveTypeRequest>,
  ): Promise<{ leaveType: LeaveType }> {
    const response = await this.axios.put<SuccessEnvelope<{ leaveType: LeaveType }>>(
      endpoints.leaveTypes.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.leaveTypes.delete(String(organizationId), String(id)));
  }
}

export class Employees {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ employees: Employee[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ employees: Employee[] }>>(
      endpoints.employees.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ employee: Employee }> {
    const response = await this.axios.get<SuccessEnvelope<{ employee: Employee }>>(
      endpoints.employees.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateEmployeeRequest,
  ): Promise<{ employee: Employee }> {
    const response = await this.axios.post<SuccessEnvelope<{ employee: Employee }>>(
      endpoints.employees.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateEmployeeRequest,
  ): Promise<{ employee: Employee }> {
    const response = await this.axios.put<SuccessEnvelope<{ employee: Employee }>>(
      endpoints.employees.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.employees.delete(String(organizationId), String(id)));
  }
}

export class Contracts {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ contracts: EmploymentContract[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ contracts: EmploymentContract[] }>>(
      endpoints.contracts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ contract: EmploymentContract }> {
    const response = await this.axios.get<SuccessEnvelope<{ contract: EmploymentContract }>>(
      endpoints.contracts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateContractRequest,
  ): Promise<{ contract: EmploymentContract }> {
    const response = await this.axios.post<SuccessEnvelope<{ contract: EmploymentContract }>>(
      endpoints.contracts.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateContractRequest,
  ): Promise<{ contract: EmploymentContract }> {
    const response = await this.axios.put<SuccessEnvelope<{ contract: EmploymentContract }>>(
      endpoints.contracts.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async terminate(organizationId: number, id: number): Promise<{ contract: EmploymentContract }> {
    const response = await this.axios.post<SuccessEnvelope<{ contract: EmploymentContract }>>(
      endpoints.contracts.terminate(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class LeaveRequests {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ leaveRequests: LeaveRequest[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ leaveRequests: LeaveRequest[] }>>(
      endpoints.leaveRequests.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ leaveRequest: LeaveRequest }> {
    const response = await this.axios.get<SuccessEnvelope<{ leaveRequest: LeaveRequest }>>(
      endpoints.leaveRequests.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateLeaveRequestRequest,
  ): Promise<{ leaveRequest: LeaveRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ leaveRequest: LeaveRequest }>>(
      endpoints.leaveRequests.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async submit(organizationId: number, id: number): Promise<{ leaveRequest: LeaveRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ leaveRequest: LeaveRequest }>>(
      endpoints.leaveRequests.submit(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async approve(organizationId: number, id: number): Promise<{ leaveRequest: LeaveRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ leaveRequest: LeaveRequest }>>(
      endpoints.leaveRequests.approve(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async refuse(organizationId: number, id: number): Promise<{ leaveRequest: LeaveRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ leaveRequest: LeaveRequest }>>(
      endpoints.leaveRequests.refuse(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async balance(
    organizationId: number,
    employeeId: number,
    leaveTypeId: number,
  ): Promise<{ balance: LeaveBalance }> {
    const response = await this.axios.get<SuccessEnvelope<{ balance: LeaveBalance }>>(
      endpoints.leaveRequests.balance(
        String(organizationId),
        String(employeeId),
        String(leaveTypeId),
      ),
    );
    return withListMeta(response.data);
  }
}

export class Attendances {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ attendances: Attendance[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ attendances: Attendance[] }>>(
      endpoints.attendances.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ attendance: Attendance }> {
    const response = await this.axios.get<SuccessEnvelope<{ attendance: Attendance }>>(
      endpoints.attendances.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async checkIn(organizationId: number, body: CheckInRequest): Promise<{ attendance: Attendance }> {
    const response = await this.axios.post<SuccessEnvelope<{ attendance: Attendance }>>(
      endpoints.attendances.checkIn(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async checkOut(
    organizationId: number,
    id: number,
    body: CheckOutRequest,
  ): Promise<{ attendance: Attendance }> {
    const response = await this.axios.post<SuccessEnvelope<{ attendance: Attendance }>>(
      endpoints.attendances.checkOut(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class Timesheets {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ timesheets: Timesheet[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ timesheets: Timesheet[] }>>(
      endpoints.timesheets.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ timesheet: Timesheet }> {
    const response = await this.axios.get<SuccessEnvelope<{ timesheet: Timesheet }>>(
      endpoints.timesheets.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateTimesheetRequest,
  ): Promise<{ timesheet: Timesheet }> {
    const response = await this.axios.post<SuccessEnvelope<{ timesheet: Timesheet }>>(
      endpoints.timesheets.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.timesheets.delete(String(organizationId), String(id)));
  }
}

export class SalaryRules {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ salaryRules: SalaryRule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ salaryRules: SalaryRule[] }>>(
      endpoints.salaryRules.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ salaryRule: SalaryRule }> {
    const response = await this.axios.get<SuccessEnvelope<{ salaryRule: SalaryRule }>>(
      endpoints.salaryRules.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateSalaryRuleRequest,
  ): Promise<{ salaryRule: SalaryRule }> {
    const response = await this.axios.post<SuccessEnvelope<{ salaryRule: SalaryRule }>>(
      endpoints.salaryRules.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateSalaryRuleRequest,
  ): Promise<{ salaryRule: SalaryRule }> {
    const response = await this.axios.put<SuccessEnvelope<{ salaryRule: SalaryRule }>>(
      endpoints.salaryRules.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.salaryRules.delete(String(organizationId), String(id)));
  }
}

export class PayrollRuns {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ payrollRuns: PayrollRun[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ payrollRuns: PayrollRun[] }>>(
      endpoints.payrollRuns.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreatePayrollRunRequest,
  ): Promise<{ run: PayrollRun }> {
    const response = await this.axios.post<SuccessEnvelope<{ run: PayrollRun }>>(
      endpoints.payrollRuns.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ run: PayrollRun; payslips: Payslip[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ run: PayrollRun; payslips: Payslip[] }>
    >(endpoints.payrollRuns.get(String(organizationId), String(id)));
    return withListMeta(response.data);
  }

  async confirm(
    organizationId: number,
    id: number,
    body: ConfirmPayrollRunRequest,
  ): Promise<{ run: PayrollRun }> {
    const response = await this.axios.post<SuccessEnvelope<{ run: PayrollRun }>>(
      endpoints.payrollRuns.confirm(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async pay(
    organizationId: number,
    id: number,
    body: PayPayrollRunRequest,
  ): Promise<{ run: PayrollRun }> {
    const response = await this.axios.post<SuccessEnvelope<{ run: PayrollRun }>>(
      endpoints.payrollRuns.pay(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async close(organizationId: number, id: number): Promise<{ run: PayrollRun }> {
    const response = await this.axios.post<SuccessEnvelope<{ run: PayrollRun }>>(
      endpoints.payrollRuns.close(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class Payslips {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ payslips: Payslip[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ payslips: Payslip[] }>>(
      endpoints.payslips.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ payslip: Payslip }> {
    const response = await this.axios.get<SuccessEnvelope<{ payslip: Payslip }>>(
      endpoints.payslips.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class Projects {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ projects: Project[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ projects: Project[] }>>(
      endpoints.projects.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ project: Project }> {
    const response = await this.axios.get<SuccessEnvelope<{ project: Project }>>(
      endpoints.projects.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateProjectRequest): Promise<{ project: Project }> {
    const response = await this.axios.post<SuccessEnvelope<{ project: Project }>>(
      endpoints.projects.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateProjectRequest,
  ): Promise<{ project: Project }> {
    const response = await this.axios.put<SuccessEnvelope<{ project: Project }>>(
      endpoints.projects.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async state(
    organizationId: number,
    id: number,
    body: SetProjectStateRequest,
  ): Promise<{ project: Project }> {
    const response = await this.axios.put<SuccessEnvelope<{ project: Project }>>(
      endpoints.projects.state(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async summary(organizationId: number, id: number): Promise<{ summary: ProjectSummary }> {
    const response = await this.axios.get<SuccessEnvelope<{ summary: ProjectSummary }>>(
      endpoints.projects.summary(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async bill(
    organizationId: number,
    id: number,
    body: BillTimeMaterialRequest,
  ): Promise<{ invoiceId: number; name: string; amountTotal: number }> {
    const response = await this.axios.post<
      SuccessEnvelope<{ invoiceId: number; name: string; amountTotal: number }>
    >(endpoints.projects.bill(String(organizationId), String(id)), body);
    return withListMeta(response.data);
  }

  tasks = {
    list: async (
      organizationId: number,
      projectId: number,
      params?: ListQuery,
    ): Promise<{ tasks: ProjectTask[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ tasks: ProjectTask[] }>>(
        endpoints.projects.tasks.list(String(organizationId), String(projectId)),
        {
          params,
        },
      );
      return withListMeta(response.data);
    },
    create: async (
      organizationId: number,
      projectId: number,
      body: CreateProjectTaskRequest,
    ): Promise<{ task: ProjectTask }> => {
      const response = await this.axios.post<SuccessEnvelope<{ task: ProjectTask }>>(
        endpoints.projects.tasks.create(String(organizationId), String(projectId)),
        body,
      );
      return withListMeta(response.data);
    },
    update: async (
      organizationId: number,
      projectId: number,
      taskId: number,
      body: UpdateProjectTaskRequest,
    ): Promise<{ task: ProjectTask }> => {
      const response = await this.axios.put<SuccessEnvelope<{ task: ProjectTask }>>(
        endpoints.projects.tasks.update(String(organizationId), String(projectId), String(taskId)),
        body,
      );
      return withListMeta(response.data);
    },
  };

  milestones = {
    list: async (
      organizationId: number,
      projectId: number,
      params?: ListQuery,
    ): Promise<{ milestones: ProjectMilestone[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ milestones: ProjectMilestone[] }>>(
        endpoints.projects.milestones.list(String(organizationId), String(projectId)),
        { params },
      );
      return withListMeta(response.data);
    },
    create: async (
      organizationId: number,
      projectId: number,
      body: CreateProjectMilestoneRequest,
    ): Promise<{ milestone: ProjectMilestone }> => {
      const response = await this.axios.post<SuccessEnvelope<{ milestone: ProjectMilestone }>>(
        endpoints.projects.milestones.create(String(organizationId), String(projectId)),
        body,
      );
      return withListMeta(response.data);
    },
    reached: async (
      organizationId: number,
      projectId: number,
      milestoneId: number,
      body: SetMilestoneReachedRequest,
    ): Promise<{ milestone: ProjectMilestone }> => {
      const response = await this.axios.put<SuccessEnvelope<{ milestone: ProjectMilestone }>>(
        endpoints.projects.milestones.reached(
          String(organizationId),
          String(projectId),
          String(milestoneId),
        ),
        body,
      );
      return withListMeta(response.data);
    },
  };
}

export class FxRates {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ fxRates: FxRate[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ fxRates: FxRate[] }>>(
      endpoints.fxRates.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ fxRate: FxRate }> {
    const response = await this.axios.get<SuccessEnvelope<{ fxRate: FxRate }>>(
      endpoints.fxRates.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, request: CreateFxRateRequest): Promise<{ fxRate: FxRate }> {
    const response = await this.axios.post<SuccessEnvelope<{ fxRate: FxRate }>>(
      endpoints.fxRates.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateFxRateRequest,
  ): Promise<{ fxRate: FxRate }> {
    const response = await this.axios.put<SuccessEnvelope<{ fxRate: FxRate }>>(
      endpoints.fxRates.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.fxRates.delete(String(organizationId), String(id)));
  }

  async resolve(
    organizationId: number,
    request: ResolveFxRateRequest,
  ): Promise<{ fxRate: ResolveFxRateResponse }> {
    const response = await this.axios.post<SuccessEnvelope<{ fxRate: ResolveFxRateResponse }>>(
      endpoints.fxRates.resolve(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class Currencies {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ currencies: Currency[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ currencies: Currency[] }>>(
      endpoints.currencies.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async get(code: string): Promise<{ currency: Currency }> {
    const response = await this.axios.get<SuccessEnvelope<{ currency: Currency }>>(
      endpoints.currencies.get(code),
    );
    return withListMeta(response.data);
  }
}

export class AccountingStandards {
  constructor(private readonly axios: AxiosInstance) {}

  async list(): Promise<{ standards: AccountingStandard[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ standards: AccountingStandard[] }>>(
      endpoints.accountingStandards.list,
    );
    return withListMeta(response.data);
  }
}

export class UnitCategories {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ categories: UnitCategory[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ categories: UnitCategory[] }>>(
      endpoints.unitGroups.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ category: UnitCategory }> {
    const response = await this.axios.get<SuccessEnvelope<{ category: UnitCategory }>>(
      endpoints.unitGroups.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: CreateUnitCategoryRequest): Promise<{ category: UnitCategory }> {
    const response = await this.axios.post<SuccessEnvelope<{ category: UnitCategory }>>(
      endpoints.unitGroups.create,
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    id: number,
    request: CreateUnitCategoryRequest,
  ): Promise<{ category: UnitCategory }> {
    const response = await this.axios.put<SuccessEnvelope<{ category: UnitCategory }>>(
      endpoints.unitGroups.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.unitGroups.delete(String(id)));
  }
}

export class Units {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ units: Unit[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ units: Unit[] }>>(
      endpoints.units.list,
      {
        params,
      },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ unit: Unit }> {
    const response = await this.axios.get<SuccessEnvelope<{ unit: Unit }>>(
      endpoints.units.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: CreateUnitRequest): Promise<{ unit: Unit }> {
    const response = await this.axios.post<SuccessEnvelope<{ unit: Unit }>>(
      endpoints.units.create,
      request,
    );
    return withListMeta(response.data);
  }

  async update(id: number, request: UpdateUnitRequest): Promise<{ unit: Unit }> {
    const response = await this.axios.put<SuccessEnvelope<{ unit: Unit }>>(
      endpoints.units.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.units.delete(String(id)));
  }

  async convert(request: ConvertUnitRequest): Promise<{ value: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ value: number }>>(
      endpoints.units.convert,
      request,
    );
    return withListMeta(response.data);
  }
}

export class Dimensions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ accounts: Dimension[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ accounts: Dimension[] }>>(
      endpoints.dimensions.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ account: Dimension }> {
    const response = await this.axios.get<SuccessEnvelope<{ account: Dimension }>>(
      endpoints.dimensions.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateDimensionRequest,
  ): Promise<{ account: Dimension }> {
    const response = await this.axios.post<SuccessEnvelope<{ account: Dimension }>>(
      endpoints.dimensions.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateDimensionRequest,
  ): Promise<{ account: Dimension }> {
    const response = await this.axios.put<SuccessEnvelope<{ account: Dimension }>>(
      endpoints.dimensions.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.dimensions.delete(String(organizationId), String(id)));
  }
}

export class PaymentTerms {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ paymentTerms: PaymentTerm[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ paymentTerms: PaymentTerm[] }>>(
      endpoints.paymentTerms.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ paymentTerm: PaymentTerm }> {
    const response = await this.axios.get<SuccessEnvelope<{ paymentTerm: PaymentTerm }>>(
      endpoints.paymentTerms.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: CreatePaymentTermRequest): Promise<{ paymentTerm: PaymentTerm }> {
    const response = await this.axios.post<SuccessEnvelope<{ paymentTerm: PaymentTerm }>>(
      endpoints.paymentTerms.create,
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    id: number,
    request: UpdatePaymentTermRequest,
  ): Promise<{ paymentTerm: PaymentTerm }> {
    const response = await this.axios.put<SuccessEnvelope<{ paymentTerm: PaymentTerm }>>(
      endpoints.paymentTerms.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.paymentTerms.delete(String(id)));
  }

  async splits(id: number, request: SplitPaymentTermRequest): Promise<{ splits: PaymentSplit[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ splits: PaymentSplit[] }>>(
      endpoints.paymentTerms.splits(String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class Accounts {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ accounts: Account[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ accounts: Account[] }>>(
      endpoints.accounts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ account: Account }> {
    const response = await this.axios.get<SuccessEnvelope<{ account: Account }>>(
      endpoints.accounts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateAccountRequest,
  ): Promise<{ account: Account }> {
    const response = await this.axios.post<SuccessEnvelope<{ account: Account }>>(
      endpoints.accounts.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateAccountRequest,
  ): Promise<{ account: Account }> {
    const response = await this.axios.put<SuccessEnvelope<{ account: Account }>>(
      endpoints.accounts.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.accounts.delete(String(organizationId), String(id)));
  }
}

export class Journals {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ journals: Journal[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ journals: Journal[] }>>(
      endpoints.journals.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ journal: Journal }> {
    const response = await this.axios.get<SuccessEnvelope<{ journal: Journal }>>(
      endpoints.journals.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateJournalRequest,
  ): Promise<{ journal: Journal }> {
    const response = await this.axios.post<SuccessEnvelope<{ journal: Journal }>>(
      endpoints.journals.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: Partial<CreateJournalRequest>,
  ): Promise<{ journal: Journal }> {
    const response = await this.axios.put<SuccessEnvelope<{ journal: Journal }>>(
      endpoints.journals.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.journals.delete(String(organizationId), String(id)));
  }
}

export class Taxes {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ taxes: Tax[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxes: Tax[] }>>(
      endpoints.taxes.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ tax: Tax }> {
    const response = await this.axios.get<SuccessEnvelope<{ tax: Tax }>>(
      endpoints.taxes.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, request: CreateTaxRequest): Promise<{ tax: Tax }> {
    const response = await this.axios.post<SuccessEnvelope<{ tax: Tax }>>(
      endpoints.taxes.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateTaxRequest,
  ): Promise<{ tax: Tax }> {
    const response = await this.axios.put<SuccessEnvelope<{ tax: Tax }>>(
      endpoints.taxes.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.taxes.delete(String(organizationId), String(id)));
  }
}

export class TaxYears {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ taxYears: TaxYear[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxYears: TaxYear[] }>>(
      endpoints.taxYears.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ taxYear: TaxYear }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxYear: TaxYear }>>(
      endpoints.taxYears.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateTaxYearRequest,
  ): Promise<{ taxYear: TaxYear }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxYear: TaxYear }>>(
      endpoints.taxYears.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateTaxYearRequest,
  ): Promise<{ taxYear: TaxYear }> {
    const response = await this.axios.put<SuccessEnvelope<{ taxYear: TaxYear }>>(
      endpoints.taxYears.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.taxYears.delete(String(organizationId), String(id)));
  }
}

export class Carriers {
  constructor(private readonly axios: AxiosInstance) {}

  async list(params?: ListQuery): Promise<{ carriers: Carrier[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ carriers: Carrier[] }>>(
      endpoints.carriers.list,
      { params },
    );
    return withListMeta(response.data);
  }

  async get(id: number): Promise<{ carrier: Carrier }> {
    const response = await this.axios.get<SuccessEnvelope<{ carrier: Carrier }>>(
      endpoints.carriers.get(String(id)),
    );
    return withListMeta(response.data);
  }

  async create(request: CreateCarrierRequest): Promise<{ carrier: Carrier }> {
    const response = await this.axios.post<SuccessEnvelope<{ carrier: Carrier }>>(
      endpoints.carriers.create,
      request,
    );
    return withListMeta(response.data);
  }

  async update(id: number, request: UpdateCarrierRequest): Promise<{ carrier: Carrier }> {
    const response = await this.axios.put<SuccessEnvelope<{ carrier: Carrier }>>(
      endpoints.carriers.update(String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(id: number): Promise<void> {
    await this.axios.delete(endpoints.carriers.delete(String(id)));
  }
}

export class Reports {
  constructor(private readonly axios: AxiosInstance) {}

  async trialBalance(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ trialBalance: TrialBalance }> {
    const response = await this.axios.get<SuccessEnvelope<{ trialBalance: TrialBalance }>>(
      endpoints.reports.trialBalance(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async aging(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ asOf: string; rows: AgingRow[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ asOf: string; rows: AgingRow[] }>>(
      endpoints.reports.aging(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async inventoryValuation(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ rows: InventoryValueRow[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ rows: InventoryValueRow[] }>>(
      endpoints.reports.inventoryValuation(String(organizationId)),
      {
        params,
      },
    );
    return withListMeta(response.data);
  }

  async profitAndLoss(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ profitAndLoss: ProfitAndLoss }> {
    const response = await this.axios.get<SuccessEnvelope<{ profitAndLoss: ProfitAndLoss }>>(
      endpoints.reports.profitAndLoss(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async balanceSheet(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ balanceSheet: BalanceSheet }> {
    const response = await this.axios.get<SuccessEnvelope<{ balanceSheet: BalanceSheet }>>(
      endpoints.reports.balanceSheet(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async cashFlow(organizationId: number, params?: ListQuery): Promise<{ cashFlow: CashFlow }> {
    const response = await this.axios.get<SuccessEnvelope<{ cashFlow: CashFlow }>>(
      endpoints.reports.cashFlow(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }
}

export class Kpis {
  constructor(private readonly axios: AxiosInstance) {}

  async sales(organizationId: number): Promise<{ kpi: SalesKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: SalesKpi }>>(
      endpoints.kpis.sales(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async pipeline(organizationId: number): Promise<{ kpi: PipelineKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: PipelineKpi }>>(
      endpoints.kpis.pipeline(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async inventory(organizationId: number): Promise<{ kpi: InventoryKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: InventoryKpi }>>(
      endpoints.kpis.inventory(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async subscription(organizationId: number): Promise<{ kpi: SubscriptionKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: SubscriptionKpi }>>(
      endpoints.kpis.subscription(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async projects(organizationId: number): Promise<{ kpi: ProjectKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: ProjectKpi }>>(
      endpoints.kpis.projects(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async payroll(organizationId: number): Promise<{ kpi: PayrollKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: PayrollKpi }>>(
      endpoints.kpis.payroll(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async finance(organizationId: number): Promise<{ kpi: FinanceKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: FinanceKpi }>>(
      endpoints.kpis.finance(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async procurement(organizationId: number): Promise<{ kpi: ProcurementKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: ProcurementKpi }>>(
      endpoints.kpis.procurement(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async manufacturing(organizationId: number): Promise<{ kpi: ManufacturingKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: ManufacturingKpi }>>(
      endpoints.kpis.manufacturing(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async arAp(organizationId: number): Promise<{ kpi: ArApKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: ArApKpi }>>(
      endpoints.kpis.arAp(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async cash(organizationId: number): Promise<{ kpi: CashKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: CashKpi }>>(
      endpoints.kpis.cash(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async inventoryRatio(organizationId: number): Promise<{ kpi: InventoryRatioKpi }> {
    const response = await this.axios.get<SuccessEnvelope<{ kpi: InventoryRatioKpi }>>(
      endpoints.kpis.inventoryRatio(String(organizationId)),
    );
    return withListMeta(response.data);
  }
}

export class FxRevaluations {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ revaluations: FxRevaluation[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ revaluations: FxRevaluation[] }>>(
      endpoints.fxRevaluations.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: RevalueRequest,
  ): Promise<{ result: RevaluationResult }> {
    const response = await this.axios.post<SuccessEnvelope<{ result: RevaluationResult }>>(
      endpoints.fxRevaluations.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class Accruals {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ accruals: Accrual[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ accruals: Accrual[] }>>(
      endpoints.accruals.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateAccrualRequest,
  ): Promise<{ result: AccrualResult }> {
    const response = await this.axios.post<SuccessEnvelope<{ result: AccrualResult }>>(
      endpoints.accruals.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async reverseDue(
    organizationId: number,
    body: ReverseAccrualsRequest,
  ): Promise<{ reversed: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ reversed: number }>>(
      endpoints.accruals.reverseDue(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class PeriodClose {
  constructor(private readonly axios: AxiosInstance) {}

  async close(organizationId: number, body: ClosePeriodRequest): Promise<{ result: CloseResult }> {
    const response = await this.axios.post<SuccessEnvelope<{ result: CloseResult }>>(
      endpoints.periodClose.close(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async yearEndRoll(
    organizationId: number,
    body: YearEndRollRequest,
  ): Promise<{ result: YearEndRollResult }> {
    const response = await this.axios.post<SuccessEnvelope<{ result: YearEndRollResult }>>(
      endpoints.periodClose.yearEndRoll(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export type HealthStatus = {
  status: string;
  service: string;
  time: string;
};

export class Health {
  constructor(private readonly axios: AxiosInstance) {}

  async check(): Promise<HealthStatus> {
    const response = await this.axios.get<SuccessEnvelope<HealthStatus>>(endpoints.health);
    return withListMeta(response.data);
  }
}

export class ApprovalRequests {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ approvalRequests: ApprovalRequest[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ approvalRequests: ApprovalRequest[] }>>(
      endpoints.approvalRequests.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateApprovalRequestRequest,
  ): Promise<{ approvalRequest: ApprovalRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ approvalRequest: ApprovalRequest }>>(
      endpoints.approvalRequests.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ approvalRequest: ApprovalRequest }> {
    const response = await this.axios.get<SuccessEnvelope<{ approvalRequest: ApprovalRequest }>>(
      endpoints.approvalRequests.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async decide(
    organizationId: number,
    id: number,
    request: DecideApprovalRequestRequest,
  ): Promise<{ approvalRequest: ApprovalRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ approvalRequest: ApprovalRequest }>>(
      endpoints.approvalRequests.decide(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class Attachments {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery & { ownerType?: string; ownerId?: number },
  ): Promise<{ attachments: Attachment[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ attachments: Attachment[] }>>(
      endpoints.attachments.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async upload(
    organizationId: number,
    request: UploadAttachmentRequest,
  ): Promise<{ attachment: Attachment }> {
    const form = new FormData();
    form.append("owner_type", request.ownerType);
    form.append("owner_id", String(request.ownerId));
    form.append("file", request.file);
    const response = await this.axios.post<SuccessEnvelope<{ attachment: Attachment }>>(
      endpoints.attachments.upload(String(organizationId)),
      form,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ attachment: Attachment }> {
    const response = await this.axios.get<SuccessEnvelope<{ attachment: Attachment }>>(
      endpoints.attachments.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async download(organizationId: number, id: number): Promise<Blob> {
    const response = await this.axios.get(
      endpoints.attachments.download(String(organizationId), String(id)),
      { responseType: "blob" },
    );
    return response.data;
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.attachments.delete(String(organizationId), String(id)));
  }
}

export class Messages {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery & { ownerType?: string; ownerId?: number },
  ): Promise<{ messages: Message[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ messages: Message[] }>>(
      endpoints.messages.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateMessageRequest,
  ): Promise<{ message: Message }> {
    const response = await this.axios.post<SuccessEnvelope<{ message: Message }>>(
      endpoints.messages.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ message: Message }> {
    const response = await this.axios.get<SuccessEnvelope<{ message: Message }>>(
      endpoints.messages.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}
