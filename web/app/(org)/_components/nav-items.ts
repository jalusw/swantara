import {
  BarChart3,
  BrainCircuit,
  Briefcase,
  Building2,
  Calculator,
  CalendarDays,
  ClipboardCheck,
  ClipboardList,
  Clock,
  CreditCard,
  DollarSign,
  Factory,
  FileText,
  Handshake,
  LayoutDashboard,
  ListTree,
  type LucideIcon,
  Package,
  Receipt,
  Settings,
  Truck,
  UserRound,
  Users,
} from "lucide-react";
import { humanizeKey } from "@/lib/utils/case";

export type OrgNavGroup = "overview" | "modules" | "settings";

export type OrgNavModule =
  | "crm"
  | "products"
  | "inventory"
  | "procurement"
  | "hr"
  | "finance"
  | "projects"
  | "quality"
  | "subscriptions"
  | "pos"
  | "service";

export type OrgNavItem = {
  key: string;
  group: OrgNavGroup;
  module?: OrgNavModule;
  href: string;
  icon: LucideIcon;
  /** Permission code (resource.action) required to see the item. */
  permission?: string;
};

export const orgNavModuleOrder: OrgNavModule[] = [
  "crm",
  "products",
  "inventory",
  "procurement",
  "hr",
  "projects",
  "quality",
  "finance",
  "subscriptions",
  "pos",
  "service",
];

export const orgNavItems: OrgNavItem[] = [
  {
    key: "dashboard",
    group: "overview",
    href: "/dashboard",
    icon: LayoutDashboard,
  },
  {
    key: "crm",
    group: "modules",
    module: "crm",
    href: "/crm",
    icon: Handshake,
    permission: "prospect.view",
  },
  {
    key: "sales",
    group: "modules",
    module: "crm",
    href: "/sale-orders",
    icon: FileText,
    permission: "sale_order.view",
  },
  {
    key: "productionOrders",
    group: "modules",
    module: "inventory",
    href: "/production-orders",
    icon: Factory,
    permission: "production_order.view",
  },
  {
    key: "planning",
    group: "modules",
    module: "inventory",
    href: "/planning",
    icon: BrainCircuit,
    permission: "production_order.view",
  },
  {
    key: "purchases",
    group: "modules",
    module: "procurement",
    href: "/purchases",
    icon: Truck,
    permission: "purchase_order.view",
  },
  {
    key: "products",
    group: "modules",
    module: "products",
    href: "/products",
    icon: Package,
    permission: "item.view",
  },
  {
    key: "stock",
    group: "modules",
    module: "inventory",
    href: "/stock",
    icon: Package,
    permission: "organization.view",
  },
  {
    key: "employees",
    group: "modules",
    module: "hr",
    href: "/employees",
    icon: UserRound,
    permission: "employee.view",
  },
  {
    key: "leaveRequests",
    group: "modules",
    module: "hr",
    href: "/leave-requests",
    icon: ClipboardList,
    permission: "leave_request.view",
  },
  {
    key: "attendance",
    group: "modules",
    module: "hr",
    href: "/attendance",
    icon: Clock,
    permission: "attendance.view",
  },
  {
    key: "projects",
    group: "modules",
    module: "projects",
    href: "/projects",
    icon: Briefcase,
    permission: "project.view",
  },
  {
    key: "payrollRuns",
    group: "modules",
    module: "hr",
    href: "/payroll-runs",
    icon: DollarSign,
    permission: "payroll_run.view",
  },
  {
    key: "accounting",
    group: "modules",
    module: "finance",
    href: "/accounting",
    icon: Calculator,
    permission: "journal_entry.view",
  },
  {
    key: "fixedAssets",
    group: "modules",
    module: "finance",
    href: "/fixed-assets",
    icon: Factory,
    permission: "fixed_asset.view",
  },
  {
    key: "quality",
    group: "modules",
    module: "quality",
    href: "/quality",
    icon: ClipboardList,
    permission: "quality_check.view",
  },
  {
    key: "subscriptions",
    group: "modules",
    module: "subscriptions",
    href: "/subscriptions",
    icon: CreditCard,
    permission: "subscription.view",
  },
  {
    key: "pos",
    group: "modules",
    module: "pos",
    href: "/pos",
    icon: Settings,
    permission: "organization.view",
  },
  {
    key: "expenses",
    group: "modules",
    module: "finance",
    href: "/expenses",
    icon: Receipt,
    permission: "expense_report.view",
  },
  {
    key: "deferrals",
    group: "modules",
    module: "finance",
    href: "/deferrals",
    icon: CalendarDays,
    permission: "journal_entry.view",
  },
  {
    key: "service",
    group: "modules",
    module: "service",
    href: "/service-orders",
    icon: ClipboardList,
    permission: "service_order.view",
  },
  {
    key: "commissions",
    group: "modules",
    module: "service",
    href: "/commission-plans",
    icon: DollarSign,
    permission: "commission_plan.view",
  },
  {
    key: "approvalRequests",
    group: "settings",
    href: "/approval-requests",
    icon: ClipboardCheck,
    permission: "organization.view",
  },
  {
    key: "activity",
    group: "settings",
    href: "/audit-logs",
    icon: ClipboardList,
    permission: "organization.view",
  },
  {
    key: "reference",
    group: "modules",
    module: "finance",
    href: "/reference",
    icon: ListTree,
    permission: "organization.view",
  },
  {
    key: "reports",
    group: "modules",
    module: "finance",
    href: "/reports",
    icon: BarChart3,
    permission: "reporting.view",
  },
  {
    key: "general",
    group: "settings",
    href: "/settings",
    icon: Settings,
    permission: "organization.view",
  },
  {
    key: "organizations",
    group: "settings",
    href: "/admin",
    icon: Building2,
    permission: "organization.view",
  },
  {
    key: "members",
    group: "settings",
    href: "/settings/members",
    icon: Users,
    permission: "member.view",
  },
];

const NAV_LABEL_OVERRIDES: Record<string, string> = {
  recipes: "Bills of materials",
};

export function navLabel(key: string): string {
  const override = NAV_LABEL_OVERRIDES[key];
  if (override) return override;
  return humanizeKey(key);
}
