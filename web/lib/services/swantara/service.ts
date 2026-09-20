import axios, { type AxiosInstance } from "axios";
import { getCsrfToken } from "@/lib/constants/cookies";
import { toCamelCase, toSnakeCase } from "@/lib/utils/case";
import {
  finishErrorProfile,
  finishResponseProfile,
  startRequestProfile,
} from "@/lib/utils/profiler";
import { isProtectedPage } from "@/lib/utils/url";
import {
  AssetCategories,
  Boms,
  Equipments,
  FixedAssets,
  Inventory,
  MaintenancePlans,
  Planning,
  PriceBooks,
  ProductCategories,
  ProductionOrders,
  Products,
  PurchaseOrders,
  PurchaseRequests,
  QualityAlerts,
  QualityChecks,
  QualityPoints,
  ServiceContracts,
  ServiceOrders,
  OutsideProcessingOrders,
  SupplierProducts,
  SupplierQuoteRequests,
} from "./catalog";
import {
  AuditLogs,
  BankStatements,
  Budgets,
  ConsolidationRuns,
  Deferrals,
  DropshipOrders,
  ExpenseCategories,
  ExpenseReports,
  IntegrationEvents,
  InterorganizationRules,
  InterorganizationTransactions,
  Invoices,
  JournalEntries,
  Payments,
  Reconciliations,
  Reminder,
  Rmas,
  SubscriptionPlans,
  Subscriptions,
  SystemConfigs,
  TaxPeriods,
  TaxReturns,
  TaxRules,
  VendorBills,
  WithholdingTaxes,
} from "./finance";
import {
  Auth,
  Contacts,
  Me,
  MemberRoles,
  Members,
  Organizations,
  Permissions,
  Users,
} from "./identity";
import { toSwantaraError } from "./mapper";
import {
  AccountingStandards,
  Accounts,
  Accruals,
  ApprovalRequests,
  Attachments,
  Attendances,
  Carriers,
  Contracts,
  Currencies,
  Departments,
  Dimensions,
  Employees,
  FxRates,
  FxRevaluations,
  Health,
  JobPositions,
  Journals,
  Kpis,
  LeaveRequests,
  LeaveTypes,
  Messages,
  PaymentTerms,
  PayrollRuns,
  Payslips,
  PeriodClose,
  Projects,
  Reports,
  SalaryRules,
  Taxes,
  TaxYears,
  Timesheets,
  UnitCategories,
  Units,
} from "./operations";
import {
  CommissionEntries,
  CommissionPlans,
  Coupons,
  CrmActivities,
  CrmLeads,
  CrmOpportunities,
  CrmPipeline,
  CrmStages,
  GiftCards,
  PosConfigs,
  PosOrders,
  PosSessions,
  SaleOrders,
  SalesGroups,
} from "./sales";

export type SwantaraServiceOptions = {
  getHeaders?: () => Promise<Record<string, string>>;
  refreshToken?: () => Promise<{
    accessToken: string;
    refreshToken: string;
  } | null>;
};

type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

let sharedRefreshPromise: Promise<TokenPair | null> | null = null;

async function defaultBrowserRefresh(): Promise<TokenPair | null> {
  if (typeof window === "undefined") return null;
  try {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    const csrfToken = getCsrfToken();
    if (csrfToken) headers["x-csrf-token"] = csrfToken;
    const response = await fetch("/api/v1/auth/refresh", {
      method: "POST",
      headers,
      credentials: "include",
      body: "{}",
    });
    if (!response.ok) return null;
    const json = (await response.json()) as {
      data?: { accessToken?: string; refreshToken?: string };
    };
    const accessToken = json.data?.accessToken;
    const refreshToken = json.data?.refreshToken;
    if (accessToken && refreshToken) return { accessToken, refreshToken };
    return null;
  } catch {
    return null;
  }
}

function queuedRefresh(refresher: () => Promise<TokenPair | null>): Promise<TokenPair | null> {
  if (!sharedRefreshPromise) {
    sharedRefreshPromise = refresher().finally(() => {
      sharedRefreshPromise = null;
    });
  }
  return sharedRefreshPromise;
}

export function resetSharedRefreshForTests(): void {
  sharedRefreshPromise = null;
}

class SwantaraService {
  private readonly axios: AxiosInstance;
  private requestTimeoutMilis = 30000;

  readonly auth: Auth;
  readonly health: Health;
  readonly me: Me;
  readonly users: Users;
  readonly organizations: Organizations;
  readonly accountingStandards: AccountingStandards;
  readonly members: Members;
  readonly memberRoles: MemberRoles;
  readonly permissions: Permissions;
  readonly fxRates: FxRates;
  readonly currencies: Currencies;
  readonly uomCategories: UnitCategories;
  readonly units: Units;
  readonly dimensions: Dimensions;
  readonly paymentTerms: PaymentTerms;
  readonly accounts: Accounts;
  readonly journals: Journals;
  readonly taxes: Taxes;
  readonly taxYears: TaxYears;
  readonly carriers: Carriers;
  readonly systemConfigs: SystemConfigs;
  readonly integrationEvents: IntegrationEvents;
  readonly auditLogs: AuditLogs;
  readonly contacts: Contacts;
  readonly productCategories: ProductCategories;
  readonly products: Products;
  readonly priceBooks: PriceBooks;
  readonly supplierProducts: SupplierProducts;
  readonly recipes: Boms;
  readonly productionOrders: ProductionOrders;
  readonly planning: Planning;
  readonly outsideProcessingOrders: OutsideProcessingOrders;
  readonly journalEntries: JournalEntries;
  readonly invoices: Invoices;
  readonly vendorBills: VendorBills;
  readonly payments: Payments;
  readonly taxPeriods: TaxPeriods;
  readonly bankStatements: BankStatements;
  readonly reconciliations: Reconciliations;
  readonly reminder: Reminder;
  readonly budgets: Budgets;
  readonly taxRules: TaxRules;
  readonly withholdingTaxes: WithholdingTaxes;
  readonly taxReturns: TaxReturns;
  readonly deferrals: Deferrals;
  readonly inventory: Inventory;
  readonly crmLeads: CrmLeads;
  readonly crmOpportunities: CrmOpportunities;
  readonly crmActivities: CrmActivities;
  readonly crmPipeline: CrmPipeline;
  readonly crmStages: CrmStages;
  readonly salesGroups: SalesGroups;
  readonly saleOrders: SaleOrders;
  readonly posConfigs: PosConfigs;
  readonly posSessions: PosSessions;
  readonly posOrders: PosOrders;
  readonly approvalRequests: ApprovalRequests;
  readonly attachments: Attachments;
  readonly messages: Messages;
  readonly qualityPoints: QualityPoints;
  readonly qualityChecks: QualityChecks;
  readonly qualityAlerts: QualityAlerts;
  readonly purchaseRequests: PurchaseRequests;
  readonly purchaseOrders: PurchaseOrders;
  readonly supplierQuoteRequests: SupplierQuoteRequests;
  readonly rmas: Rmas;
  readonly departments: Departments;
  readonly jobPositions: JobPositions;
  readonly leaveTypes: LeaveTypes;
  readonly employees: Employees;
  readonly contracts: Contracts;
  readonly leaveRequests: LeaveRequests;
  readonly attendances: Attendances;
  readonly timesheets: Timesheets;
  readonly salaryRules: SalaryRules;
  readonly payrollRuns: PayrollRuns;
  readonly payslips: Payslips;
  readonly projects: Projects;
  readonly expenseCategories: ExpenseCategories;
  readonly expenseReports: ExpenseReports;
  readonly assetCategories: AssetCategories;
  readonly fixedAssets: FixedAssets;
  readonly subscriptionPlans: SubscriptionPlans;
  readonly subscriptions: Subscriptions;
  readonly commissionPlans: CommissionPlans;
  readonly commissionEntries: CommissionEntries;
  readonly giftCards: GiftCards;
  readonly coupons: Coupons;
  readonly equipments: Equipments;
  readonly serviceContracts: ServiceContracts;
  readonly serviceOrders: ServiceOrders;
  readonly maintenancePlans: MaintenancePlans;
  readonly dropshipOrders: DropshipOrders;
  readonly interorganizationRules: InterorganizationRules;
  readonly interorganizationTransactions: InterorganizationTransactions;
  readonly consolidationRuns: ConsolidationRuns;
  readonly reports: Reports;
  readonly kpis: Kpis;
  readonly fxRevaluations: FxRevaluations;
  readonly accruals: Accruals;
  readonly periodClose: PeriodClose;

  constructor(url: string, options?: SwantaraServiceOptions) {
    this.axios = axios.create({
      baseURL: url,
      timeout: this.requestTimeoutMilis,
      withCredentials: true,
    });

    this.axios.interceptors.request.use((config) => {
      const csrfToken = getCsrfToken();
      if (csrfToken) {
        config.headers.set("x-csrf-token", csrfToken);
      }
      return config;
    });

    if (options?.getHeaders) {
      const getHeaders = options.getHeaders;
      this.axios.interceptors.request.use(async (config) => {
        config.headers.set(await getHeaders());
        return config;
      }, Promise.reject);
    }

    this.axios.interceptors.request.use((config) => {
      startRequestProfile(config);
      config.data = toSnakeCase(config.data);
      return config;
    }, Promise.reject);

    this.axios.interceptors.response.use(
      (response) => {
        response.data = toCamelCase(response.data);
        finishResponseProfile(response);
        return response;
      },
      (error) => {
        const is401 = error?.response?.status === 401;
        const isAuthEndpoint =
          error?.config?.url?.includes("/auth/login") ||
          error?.config?.url?.includes("/auth/refresh");
        const config = error.config as Record<string, unknown>;
        const isBrowser = typeof window !== "undefined";
        const canRefresh = is401 && !isAuthEndpoint && isBrowser && !config._retry;
        const getLoginPath = () => {
          return "/login";
        };
        if (canRefresh) {
          const refresher = options?.refreshToken ?? defaultBrowserRefresh;
          return queuedRefresh(refresher).then((newTokens) => {
            if (newTokens) {
              config._retry = true;
              const retryHeaders = error.config.headers as
                | { set?: (name: string, value: string) => void; Authorization?: string }
                | undefined;
              if (typeof retryHeaders?.set === "function") {
                retryHeaders.set("Authorization", `Bearer ${newTokens.accessToken}`);
              } else if (retryHeaders) {
                retryHeaders.Authorization = `Bearer ${newTokens.accessToken}`;
              }
              return this.axios.request(error.config);
            }
            finishErrorProfile(error);
            if (isBrowser && isProtectedPage(window.location.pathname)) {
              window.location.href = getLoginPath();
            }
            return Promise.reject(toSwantaraError(error));
          });
        }
        finishErrorProfile(error);
        if (
          is401 &&
          !isAuthEndpoint &&
          typeof window !== "undefined" &&
          isProtectedPage(window.location.pathname)
        ) {
          window.location.href = getLoginPath();
        }
        return Promise.reject(toSwantaraError(error));
      },
    );

    this.auth = new Auth(this.axios);
    this.health = new Health(this.axios);
    this.me = new Me(this.axios);
    this.users = new Users(this.axios);
    this.organizations = new Organizations(this.axios);
    this.accountingStandards = new AccountingStandards(this.axios);
    this.members = new Members(this.axios);
    this.memberRoles = new MemberRoles(this.axios);
    this.permissions = new Permissions(this.axios);
    this.fxRates = new FxRates(this.axios);
    this.currencies = new Currencies(this.axios);
    this.uomCategories = new UnitCategories(this.axios);
    this.units = new Units(this.axios);
    this.dimensions = new Dimensions(this.axios);
    this.paymentTerms = new PaymentTerms(this.axios);
    this.accounts = new Accounts(this.axios);
    this.journals = new Journals(this.axios);
    this.taxes = new Taxes(this.axios);
    this.taxYears = new TaxYears(this.axios);
    this.carriers = new Carriers(this.axios);
    this.systemConfigs = new SystemConfigs(this.axios);
    this.integrationEvents = new IntegrationEvents(this.axios);
    this.auditLogs = new AuditLogs(this.axios);
    this.contacts = new Contacts(this.axios);
    this.productCategories = new ProductCategories(this.axios);
    this.products = new Products(this.axios);
    this.priceBooks = new PriceBooks(this.axios);
    this.supplierProducts = new SupplierProducts(this.axios);
    this.recipes = new Boms(this.axios);
    this.productionOrders = new ProductionOrders(this.axios);
    this.planning = new Planning(this.axios);
    this.outsideProcessingOrders = new OutsideProcessingOrders(this.axios);
    this.journalEntries = new JournalEntries(this.axios);
    this.invoices = new Invoices(this.axios);
    this.vendorBills = new VendorBills(this.axios);
    this.payments = new Payments(this.axios);
    this.taxPeriods = new TaxPeriods(this.axios);
    this.bankStatements = new BankStatements(this.axios);
    this.reconciliations = new Reconciliations(this.axios);
    this.reminder = new Reminder(this.axios);
    this.budgets = new Budgets(this.axios);
    this.taxRules = new TaxRules(this.axios);
    this.withholdingTaxes = new WithholdingTaxes(this.axios);
    this.taxReturns = new TaxReturns(this.axios);
    this.deferrals = new Deferrals(this.axios);
    this.inventory = new Inventory(this.axios);
    this.crmLeads = new CrmLeads(this.axios);
    this.crmOpportunities = new CrmOpportunities(this.axios);
    this.crmActivities = new CrmActivities(this.axios);
    this.crmPipeline = new CrmPipeline(this.axios);
    this.crmStages = new CrmStages(this.axios);
    this.salesGroups = new SalesGroups(this.axios);
    this.saleOrders = new SaleOrders(this.axios);
    this.posConfigs = new PosConfigs(this.axios);
    this.posSessions = new PosSessions(this.axios);
    this.posOrders = new PosOrders(this.axios);
    this.approvalRequests = new ApprovalRequests(this.axios);
    this.attachments = new Attachments(this.axios);
    this.messages = new Messages(this.axios);
    this.qualityPoints = new QualityPoints(this.axios);
    this.qualityChecks = new QualityChecks(this.axios);
    this.qualityAlerts = new QualityAlerts(this.axios);
    this.purchaseRequests = new PurchaseRequests(this.axios);
    this.purchaseOrders = new PurchaseOrders(this.axios);
    this.supplierQuoteRequests = new SupplierQuoteRequests(this.axios);
    this.rmas = new Rmas(this.axios);
    this.departments = new Departments(this.axios);
    this.jobPositions = new JobPositions(this.axios);
    this.leaveTypes = new LeaveTypes(this.axios);
    this.employees = new Employees(this.axios);
    this.contracts = new Contracts(this.axios);
    this.leaveRequests = new LeaveRequests(this.axios);
    this.attendances = new Attendances(this.axios);
    this.timesheets = new Timesheets(this.axios);
    this.salaryRules = new SalaryRules(this.axios);
    this.payrollRuns = new PayrollRuns(this.axios);
    this.payslips = new Payslips(this.axios);
    this.projects = new Projects(this.axios);
    this.expenseCategories = new ExpenseCategories(this.axios);
    this.expenseReports = new ExpenseReports(this.axios);
    this.assetCategories = new AssetCategories(this.axios);
    this.fixedAssets = new FixedAssets(this.axios);
    this.subscriptionPlans = new SubscriptionPlans(this.axios);
    this.subscriptions = new Subscriptions(this.axios);
    this.commissionPlans = new CommissionPlans(this.axios);
    this.commissionEntries = new CommissionEntries(this.axios);
    this.giftCards = new GiftCards(this.axios);
    this.coupons = new Coupons(this.axios);
    this.equipments = new Equipments(this.axios);
    this.serviceContracts = new ServiceContracts(this.axios);
    this.serviceOrders = new ServiceOrders(this.axios);
    this.maintenancePlans = new MaintenancePlans(this.axios);
    this.dropshipOrders = new DropshipOrders(this.axios);
    this.interorganizationRules = new InterorganizationRules(this.axios);
    this.interorganizationTransactions = new InterorganizationTransactions(this.axios);
    this.consolidationRuns = new ConsolidationRuns(this.axios);
    this.reports = new Reports(this.axios);
    this.kpis = new Kpis(this.axios);
    this.fxRevaluations = new FxRevaluations(this.axios);
    this.accruals = new Accruals(this.axios);
    this.periodClose = new PeriodClose(this.axios);
  }
}

export { SwantaraService };

let serviceInstance: SwantaraService | null = null;

export function getSwantaraService(): SwantaraService {
  if (!serviceInstance) {
    serviceInstance = new SwantaraService("/");
  }
  return serviceInstance;
}
