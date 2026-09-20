"use client";

import type {
  ArApKpi,
  CashKpi,
  FinanceKpi,
  InventoryKpi,
  InventoryRatioKpi,
  ManufacturingKpi,
  PayrollKpi,
  PipelineKpi,
  ProcurementKpi,
  ProjectKpi,
  SalesKpi,
  SubscriptionKpi,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { useOrgListQuery } from "./use-org-query";

export function useFinanceKpi() {
  return useOrgListQuery<{ kpi: FinanceKpi }, Record<string, never>>(
    "financeKpi",
    (organizationId) => getSwantaraService().kpis.finance(organizationId),
  );
}

export function usePayrollKpi() {
  return useOrgListQuery<{ kpi: PayrollKpi }, Record<string, never>>(
    "payrollKpi",
    (organizationId) => getSwantaraService().kpis.payroll(organizationId),
  );
}

export function useCashKpi() {
  return useOrgListQuery<{ kpi: CashKpi }, Record<string, never>>("cashKpi", (organizationId) =>
    getSwantaraService().kpis.cash(organizationId),
  );
}

export function useSalesKpi() {
  return useOrgListQuery<{ kpi: SalesKpi }, Record<string, never>>("salesKpi", (organizationId) =>
    getSwantaraService().kpis.sales(organizationId),
  );
}

export function usePipelineKpi() {
  return useOrgListQuery<{ kpi: PipelineKpi }, Record<string, never>>(
    "pipelineKpi",
    (organizationId) => getSwantaraService().kpis.pipeline(organizationId),
  );
}

export function useInventoryKpi() {
  return useOrgListQuery<{ kpi: InventoryKpi }, Record<string, never>>(
    "inventoryKpi",
    (organizationId) => getSwantaraService().kpis.inventory(organizationId),
  );
}

export function useSubscriptionKpi() {
  return useOrgListQuery<{ kpi: SubscriptionKpi }, Record<string, never>>(
    "subscriptionKpi",
    (organizationId) => getSwantaraService().kpis.subscription(organizationId),
  );
}

export function useProjectKpi() {
  return useOrgListQuery<{ kpi: ProjectKpi }, Record<string, never>>(
    "projectKpi",
    (organizationId) => getSwantaraService().kpis.projects(organizationId),
  );
}

export function useProcurementKpi() {
  return useOrgListQuery<{ kpi: ProcurementKpi }, Record<string, never>>(
    "procurementKpi",
    (organizationId) => getSwantaraService().kpis.procurement(organizationId),
  );
}

export function useManufacturingKpi() {
  return useOrgListQuery<{ kpi: ManufacturingKpi }, Record<string, never>>(
    "manufacturingKpi",
    (organizationId) => getSwantaraService().kpis.manufacturing(organizationId),
  );
}

export function useArApKpi() {
  return useOrgListQuery<{ kpi: ArApKpi }, Record<string, never>>("arApKpi", (organizationId) =>
    getSwantaraService().kpis.arAp(organizationId),
  );
}

export function useInventoryRatioKpi() {
  return useOrgListQuery<{ kpi: InventoryRatioKpi }, Record<string, never>>(
    "inventoryRatioKpi",
    (organizationId) => getSwantaraService().kpis.inventoryRatio(organizationId),
  );
}
