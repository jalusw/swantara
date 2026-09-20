import type { AxiosInstance } from "axios";
import { endpoints } from "./endpoints";
import type {
  AccountPartialReconcile,
  AuditLog,
  BankStatement,
  Budget,
  BudgetVarianceLine,
  ConsolidatedBalance,
  ConsolidationElimination,
  ConsolidationRun,
  CreateBankStatementRequest,
  CreateBudgetRequest,
  CreateConsolidationRunRequest,
  CreateCreditNoteRequest,
  CreateDeferralScheduleRequest,
  CreateDropshipOrderRequest,
  CreateExpenseCategoryRequest,
  CreateExpenseReportRequest,
  CreateInvoiceRequest,
  CreatePaymentRequest,
  CreateRmaRequest,
  CreateSubscriptionPlanRequest,
  CreateSubscriptionRequest,
  CreateSystemConfigRequest,
  CreateTaxPeriodRequest,
  CreateTaxReturnRequest,
  CreateTaxRuleRequest,
  CreateWithholdingTaxRequest,
  DeferralLine,
  DeferralSchedule,
  DropshipLink,
  ExpenseCategory,
  ExpenseReport,
  GenerateReminderRequest,
  IntegrationEvent,
  InterorganizationRule,
  InterorganizationTransaction,
  Invoice,
  InvoiceSummary,
  JournalEntry,
  ListQuery,
  MirrorSaleOrderRequest,
  Payment,
  PaymentSummary,
  PostingLineRequest,
  PurchaseOrder,
  ReceiveDropshipOrderRequest,
  ReceiveRmaRequest,
  RecognizeDeferralsRequest,
  ReconcileRequest,
  RefundRmaRequest,
  ReminderAction,
  ResolveTaxRuleData,
  ResolveTaxRuleRequest,
  ReverseJournalEntryRequest,
  Rma,
  RmaLine,
  Subscription,
  SubscriptionMetrics,
  SubscriptionPlan,
  SuccessEnvelope,
  SystemConfig,
  TaxPeriod,
  TaxReturn,
  TaxRule,
  UpdateExpenseCategoryRequest,
  UpdateSubscriptionPlanRequest,
  UpdateSystemConfigRequest,
  UpsertInterorganizationRuleRequest,
  WithholdingTax,
  WithholdRequest,
} from "./types";
import { withListMeta } from "./types";

export type CreateJournalEntryRequest = {
  organizationId: number | null;
  journalId: number;
  date: string;
  ref: string;
  originType: string;
  originId: number;
  description: string;
  lines: PostingLineRequest[];
};

export class JournalEntries {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ movements: JournalEntry[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ movements: JournalEntry[] }>>(
      endpoints.journalEntries.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ movement: JournalEntry }> {
    const response = await this.axios.get<SuccessEnvelope<{ movement: JournalEntry }>>(
      endpoints.journalEntries.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateJournalEntryRequest,
  ): Promise<{ movement: JournalEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ movement: JournalEntry }>>(
      endpoints.journalEntries.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async reverse(
    organizationId: number,
    id: number,
    body: ReverseJournalEntryRequest,
  ): Promise<{ movement: JournalEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ movement: JournalEntry }>>(
      endpoints.journalEntries.reverse(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class Invoices {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ invoices: InvoiceSummary[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ invoices: InvoiceSummary[] }>>(
      endpoints.invoices.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ invoice: Invoice }> {
    const response = await this.axios.get<SuccessEnvelope<{ invoice: Invoice }>>(
      endpoints.invoices.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateInvoiceRequest): Promise<{ invoice: Invoice }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: Invoice }>>(
      endpoints.invoices.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async creditNote(
    organizationId: number,
    id: number,
    body: CreateCreditNoteRequest,
  ): Promise<{ invoice: Invoice }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: Invoice }>>(
      endpoints.invoices.creditNote(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class VendorBills {
  constructor(private readonly axios: AxiosInstance) {}

  async create(
    organizationId: number,
    body: CreateInvoiceRequest,
  ): Promise<{ invoice: InvoiceSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: InvoiceSummary }>>(
      endpoints.supplierBills.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async creditNote(
    organizationId: number,
    id: number,
    body: CreateCreditNoteRequest,
  ): Promise<{ invoice: InvoiceSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: InvoiceSummary }>>(
      endpoints.supplierBills.creditNote(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class Payments {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ payments: Payment[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ payments: Payment[] }>>(
      endpoints.payments.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ payment: Payment }> {
    const response = await this.axios.get<SuccessEnvelope<{ payment: Payment }>>(
      endpoints.payments.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreatePaymentRequest,
  ): Promise<{ payment: PaymentSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ payment: PaymentSummary }>>(
      endpoints.payments.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async createOutbound(
    organizationId: number,
    body: CreatePaymentRequest,
  ): Promise<{ payment: PaymentSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ payment: PaymentSummary }>>(
      endpoints.payments.createOutbound(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class TaxPeriods {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ taxPeriods: TaxPeriod[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxPeriods: TaxPeriod[] }>>(
      endpoints.taxPeriods.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ taxPeriod: TaxPeriod }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxPeriod: TaxPeriod }>>(
      endpoints.taxPeriods.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateTaxPeriodRequest,
  ): Promise<{ taxPeriod: TaxPeriod }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxPeriod: TaxPeriod }>>(
      endpoints.taxPeriods.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async close(organizationId: number, id: number): Promise<{ taxPeriod: TaxPeriod }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxPeriod: TaxPeriod }>>(
      endpoints.taxPeriods.close(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async lock(organizationId: number, id: number): Promise<{ taxPeriod: TaxPeriod }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxPeriod: TaxPeriod }>>(
      endpoints.taxPeriods.lock(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async open(organizationId: number, id: number): Promise<{ taxPeriod: TaxPeriod }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxPeriod: TaxPeriod }>>(
      endpoints.taxPeriods.open(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class BankStatements {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ bankStatements: BankStatement[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ bankStatements: BankStatement[] }>>(
      endpoints.bankStatements.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ bankStatement: BankStatement }> {
    const response = await this.axios.get<SuccessEnvelope<{ bankStatement: BankStatement }>>(
      endpoints.bankStatements.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateBankStatementRequest,
  ): Promise<{ bankStatement: BankStatement }> {
    const response = await this.axios.post<SuccessEnvelope<{ bankStatement: BankStatement }>>(
      endpoints.bankStatements.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async match(
    organizationId: number,
    id: number,
    body: ReconcileRequest,
  ): Promise<{ unreconciled: AccountPartialReconcile[] }> {
    const response = await this.axios.post<
      SuccessEnvelope<{ unreconciled: AccountPartialReconcile[] }>
    >(endpoints.bankStatements.match(String(organizationId), String(id)), body);
    return withListMeta(response.data);
  }
}

export class Reconciliations {
  constructor(private readonly axios: AxiosInstance) {}

  async create(
    organizationId: number,
    body: ReconcileRequest,
  ): Promise<{
    partialReconcile: AccountPartialReconcile[];
    reconciled: AccountPartialReconcile[];
  }> {
    const response = await this.axios.post<
      SuccessEnvelope<{
        partialReconcile: AccountPartialReconcile[];
        reconciled: AccountPartialReconcile[];
      }>
    >(endpoints.reconciliations.create(String(organizationId)), body);
    return withListMeta(response.data);
  }
}

export class Reminder {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ actions: ReminderAction[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ actions: ReminderAction[] }>>(
      endpoints.reminder.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async generate(
    organizationId: number,
    body: GenerateReminderRequest,
  ): Promise<{ actions: ReminderAction[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ actions: ReminderAction[] }>>(
      endpoints.reminder.generate(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class Budgets {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ budgets: Budget[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ budgets: Budget[] }>>(
      endpoints.budgets.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ budget: Budget }> {
    const response = await this.axios.get<SuccessEnvelope<{ budget: Budget }>>(
      endpoints.budgets.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateBudgetRequest): Promise<{ budget: Budget }> {
    const response = await this.axios.post<SuccessEnvelope<{ budget: Budget }>>(
      endpoints.budgets.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async variance(organizationId: number, id: number): Promise<{ variance: BudgetVarianceLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ variance: BudgetVarianceLine[] }>>(
      endpoints.budgets.variance(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class TaxRules {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ taxRules: TaxRule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxRules: TaxRule[] }>>(
      endpoints.taxRules.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ taxRule: TaxRule }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxRule: TaxRule }>>(
      endpoints.taxRules.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateTaxRuleRequest): Promise<{ taxRule: TaxRule }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxRule: TaxRule }>>(
      endpoints.taxRules.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async resolve(
    organizationId: number,
    id: number,
    body: ResolveTaxRuleRequest,
  ): Promise<{ data: ResolveTaxRuleData }> {
    const response = await this.axios.post<SuccessEnvelope<{ data: ResolveTaxRuleData }>>(
      endpoints.taxRules.resolve(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class WithholdingTaxes {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ withholdingTaxes: WithholdingTax[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ withholdingTaxes: WithholdingTax[] }>>(
      endpoints.withholdingTaxes.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateWithholdingTaxRequest,
  ): Promise<{ withholdingTax: WithholdingTax }> {
    const response = await this.axios.post<SuccessEnvelope<{ withholdingTax: WithholdingTax }>>(
      endpoints.withholdingTaxes.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async apply(organizationId: number, body: WithholdRequest): Promise<{ moveId: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ moveId: number }>>(
      endpoints.withholdingTaxes.apply(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class TaxReturns {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ taxReturns: TaxReturn[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxReturns: TaxReturn[] }>>(
      endpoints.taxReturns.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ taxReturn: TaxReturn }> {
    const response = await this.axios.get<SuccessEnvelope<{ taxReturn: TaxReturn }>>(
      endpoints.taxReturns.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateTaxReturnRequest,
  ): Promise<{ taxReturn: TaxReturn }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxReturn: TaxReturn }>>(
      endpoints.taxReturns.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async file(organizationId: number, id: number): Promise<{ taxReturn: TaxReturn }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxReturn: TaxReturn }>>(
      endpoints.taxReturns.file(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async pay(organizationId: number, id: number): Promise<{ taxReturn: TaxReturn }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxReturn: TaxReturn }>>(
      endpoints.taxReturns.pay(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async open(organizationId: number, id: number): Promise<{ taxReturn: TaxReturn }> {
    const response = await this.axios.post<SuccessEnvelope<{ taxReturn: TaxReturn }>>(
      endpoints.taxReturns.open(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class Deferrals {
  constructor(private readonly axios: AxiosInstance) {}

  async recognize(
    organizationId: number,
    body: RecognizeDeferralsRequest,
  ): Promise<{ posted: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ posted: number }>>(
      endpoints.deferrals.recognize(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ schedules: DeferralSchedule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ schedules: DeferralSchedule[] }>>(
      endpoints.deferrals.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateDeferralScheduleRequest,
  ): Promise<{ schedule: DeferralSchedule }> {
    const response = await this.axios.post<SuccessEnvelope<{ schedule: DeferralSchedule }>>(
      endpoints.deferrals.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ schedule: DeferralSchedule }> {
    const response = await this.axios.get<SuccessEnvelope<{ schedule: DeferralSchedule }>>(
      endpoints.deferrals.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async lines(organizationId: number, id: number): Promise<{ lines: DeferralLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ lines: DeferralLine[] }>>(
      endpoints.deferrals.lines(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class ExpenseCategories {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ categories: ExpenseCategory[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ categories: ExpenseCategory[] }>>(
      endpoints.expenseCategories.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ category: ExpenseCategory }> {
    const response = await this.axios.get<SuccessEnvelope<{ category: ExpenseCategory }>>(
      endpoints.expenseCategories.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateExpenseCategoryRequest,
  ): Promise<{ category: ExpenseCategory }> {
    const response = await this.axios.post<SuccessEnvelope<{ category: ExpenseCategory }>>(
      endpoints.expenseCategories.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateExpenseCategoryRequest,
  ): Promise<{ category: ExpenseCategory }> {
    const response = await this.axios.put<SuccessEnvelope<{ category: ExpenseCategory }>>(
      endpoints.expenseCategories.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.expenseCategories.delete(String(organizationId), String(id)));
  }
}

export class ExpenseReports {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ reports: ExpenseReport[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ reports: ExpenseReport[] }>>(
      endpoints.expenseReports.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ report: ExpenseReport }> {
    const response = await this.axios.get<SuccessEnvelope<{ report: ExpenseReport }>>(
      endpoints.expenseReports.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateExpenseReportRequest,
  ): Promise<{ report: ExpenseReport }> {
    const response = await this.axios.post<SuccessEnvelope<{ report: ExpenseReport }>>(
      endpoints.expenseReports.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async submit(organizationId: number, id: number): Promise<void> {
    await this.axios.post(endpoints.expenseReports.submit(String(organizationId), String(id)));
  }

  async approve(organizationId: number, id: number): Promise<void> {
    await this.axios.post(endpoints.expenseReports.approve(String(organizationId), String(id)));
  }

  async refuse(organizationId: number, id: number): Promise<void> {
    await this.axios.post(endpoints.expenseReports.refuse(String(organizationId), String(id)));
  }

  async post(organizationId: number, id: number): Promise<void> {
    await this.axios.post(endpoints.expenseReports.post(String(organizationId), String(id)));
  }

  async reimburse(organizationId: number, id: number): Promise<void> {
    await this.axios.post(endpoints.expenseReports.reimburse(String(organizationId), String(id)));
  }

  async bill(organizationId: number, id: number): Promise<{ invoiceId: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoiceId: number }>>(
      endpoints.expenseReports.bill(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class Rmas {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ rmas: Rma[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ rmas: Rma[] }>>(
      endpoints.rmas.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ rma: Rma; lines: RmaLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ rma: Rma; lines: RmaLine[] }>>(
      endpoints.rmas.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateRmaRequest,
  ): Promise<{ rma: Rma; lines: RmaLine[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma; lines: RmaLine[] }>>(
      endpoints.rmas.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async confirm(organizationId: number, id: number): Promise<{ rma: Rma }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma }>>(
      endpoints.rmas.confirm(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async receive(
    organizationId: number,
    id: number,
    request: ReceiveRmaRequest,
  ): Promise<{ rma: Rma }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma }>>(
      endpoints.rmas.receive(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async refund(
    organizationId: number,
    id: number,
    request: RefundRmaRequest,
  ): Promise<{ rma: Rma }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma }>>(
      endpoints.rmas.refund(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async done(organizationId: number, id: number): Promise<{ rma: Rma }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma }>>(
      endpoints.rmas.done(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ rma: Rma }> {
    const response = await this.axios.post<SuccessEnvelope<{ rma: Rma }>>(
      endpoints.rmas.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class DropshipOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async create(
    organizationId: number,
    body: CreateDropshipOrderRequest,
  ): Promise<{ purchaseOrder: PurchaseOrder; links: DropshipLink[] }> {
    const response = await this.axios.post<
      SuccessEnvelope<{ purchaseOrder: PurchaseOrder; links: DropshipLink[] }>
    >(endpoints.dropshipOrders.create(String(organizationId)), body);
    return withListMeta(response.data);
  }

  async receive(
    organizationId: number,
    id: number,
    body: ReceiveDropshipOrderRequest,
  ): Promise<{ purchaseOrder: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ purchaseOrder: PurchaseOrder }>>(
      endpoints.dropshipOrders.receive(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class InterorganizationRules {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ rules: InterorganizationRule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ rules: InterorganizationRule[] }>>(
      endpoints.interorganizationRules.list(String(organizationId)),
      {
        params,
      },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: UpsertInterorganizationRuleRequest,
  ): Promise<{ rule: InterorganizationRule }> {
    const response = await this.axios.post<SuccessEnvelope<{ rule: InterorganizationRule }>>(
      endpoints.interorganizationRules.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpsertInterorganizationRuleRequest,
  ): Promise<{ rule: InterorganizationRule }> {
    const response = await this.axios.put<SuccessEnvelope<{ rule: InterorganizationRule }>>(
      endpoints.interorganizationRules.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(
      endpoints.interorganizationRules.delete(String(organizationId), String(id)),
    );
  }
}

export class InterorganizationTransactions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ transactions: InterorganizationTransaction[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ transactions: InterorganizationTransaction[] }>
    >(endpoints.interorganizationTransactions.list(String(organizationId)), {
      params,
    });
    return withListMeta(response.data);
  }

  async mirrorSaleOrder(
    organizationId: number,
    id: number,
    body: MirrorSaleOrderRequest,
  ): Promise<{
    purchaseOrder: PurchaseOrder;
    interorganizationTransaction: InterorganizationTransaction;
  }> {
    const response = await this.axios.post<
      SuccessEnvelope<{
        purchaseOrder: PurchaseOrder;
        interorganizationTransaction: InterorganizationTransaction;
      }>
    >(
      endpoints.interorganizationTransactions.mirrorSaleOrder(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class ConsolidationRuns {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ consolidationRuns: ConsolidationRun[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ consolidationRuns: ConsolidationRun[] }>
    >(endpoints.consolidationRuns.list(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateConsolidationRunRequest,
  ): Promise<{ consolidationRun: ConsolidationRun }> {
    const response = await this.axios.post<SuccessEnvelope<{ consolidationRun: ConsolidationRun }>>(
      endpoints.consolidationRuns.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async get(
    organizationId: number,
    id: number,
  ): Promise<{
    consolidationRun: ConsolidationRun;
    eliminations: ConsolidationElimination[];
  }> {
    const response = await this.axios.get<
      SuccessEnvelope<{
        consolidationRun: ConsolidationRun;
        eliminations: ConsolidationElimination[];
      }>
    >(endpoints.consolidationRuns.get(String(organizationId), String(id)));
    return withListMeta(response.data);
  }

  async run(
    organizationId: number,
    id: number,
  ): Promise<{
    consolidationRun: ConsolidationRun;
    memberBalances: ConsolidatedBalance[];
    eliminations: ConsolidationElimination[];
  }> {
    const response = await this.axios.post<
      SuccessEnvelope<{
        consolidationRun: ConsolidationRun;
        memberBalances: ConsolidatedBalance[];
        eliminations: ConsolidationElimination[];
      }>
    >(endpoints.consolidationRuns.run(String(organizationId), String(id)));
    return withListMeta(response.data);
  }
}

export class SubscriptionPlans {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ plans: SubscriptionPlan[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ plans: SubscriptionPlan[] }>>(
      endpoints.subscriptionPlans.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ plan: SubscriptionPlan }> {
    const response = await this.axios.get<SuccessEnvelope<{ plan: SubscriptionPlan }>>(
      endpoints.subscriptionPlans.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateSubscriptionPlanRequest,
  ): Promise<{ plan: SubscriptionPlan }> {
    const response = await this.axios.post<SuccessEnvelope<{ plan: SubscriptionPlan }>>(
      endpoints.subscriptionPlans.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateSubscriptionPlanRequest,
  ): Promise<{ plan: SubscriptionPlan }> {
    const response = await this.axios.put<SuccessEnvelope<{ plan: SubscriptionPlan }>>(
      endpoints.subscriptionPlans.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.subscriptionPlans.delete(String(organizationId), String(id)));
  }
}

export class Subscriptions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ subscriptions: Subscription[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ subscriptions: Subscription[] }>>(
      endpoints.subscriptions.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateSubscriptionRequest,
  ): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async metrics(organizationId: number): Promise<{ metrics: SubscriptionMetrics }> {
    const response = await this.axios.get<SuccessEnvelope<{ metrics: SubscriptionMetrics }>>(
      endpoints.subscriptions.metrics(String(organizationId)),
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.get<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async activate(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.activate(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async pause(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.pause(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async resume(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.resume(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async churn(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.churn(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async close(organizationId: number, id: number): Promise<{ subscription: Subscription }> {
    const response = await this.axios.post<SuccessEnvelope<{ subscription: Subscription }>>(
      endpoints.subscriptions.close(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class SystemConfigs {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ systemConfigs: SystemConfig[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ systemConfigs: SystemConfig[] }>>(
      endpoints.systemConfigs.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateSystemConfigRequest,
  ): Promise<{ systemConfig: SystemConfig }> {
    const response = await this.axios.post<SuccessEnvelope<{ systemConfig: SystemConfig }>>(
      endpoints.systemConfigs.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ systemConfig: SystemConfig }> {
    const response = await this.axios.get<SuccessEnvelope<{ systemConfig: SystemConfig }>>(
      endpoints.systemConfigs.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateSystemConfigRequest,
  ): Promise<{ systemConfig: SystemConfig }> {
    const response = await this.axios.put<SuccessEnvelope<{ systemConfig: SystemConfig }>>(
      endpoints.systemConfigs.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.systemConfigs.delete(String(organizationId), String(id)));
  }
}

export class IntegrationEvents {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ integrationEvents: IntegrationEvent[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ integrationEvents: IntegrationEvent[] }>
    >(endpoints.integrationEvents.list(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ integrationEvent: IntegrationEvent }> {
    const response = await this.axios.get<SuccessEnvelope<{ integrationEvent: IntegrationEvent }>>(
      endpoints.integrationEvents.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async dispatch(
    organizationId: number,
    id: number,
  ): Promise<{ integrationEvent: IntegrationEvent }> {
    const response = await this.axios.post<SuccessEnvelope<{ integrationEvent: IntegrationEvent }>>(
      endpoints.integrationEvents.dispatch(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class AuditLogs {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ auditLogs: AuditLog[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ auditLogs: AuditLog[] }>>(
      endpoints.auditLogs.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ auditLog: AuditLog }> {
    const response = await this.axios.get<SuccessEnvelope<{ auditLog: AuditLog }>>(
      endpoints.auditLogs.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}
