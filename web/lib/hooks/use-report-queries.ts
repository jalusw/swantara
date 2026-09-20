"use client";

import type {
  AgingRow,
  BalanceSheet,
  CashFlow,
  InventoryValueRow,
  ListQuery,
  ProfitAndLoss,
  TrialBalance,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { useOrgListQuery } from "./use-org-query";

export function useTrialBalance(params?: ListQuery) {
  return useOrgListQuery<{ trialBalance: TrialBalance }, ListQuery>(
    "trialBalance",
    (organizationId, p) => getSwantaraService().reports.trialBalance(organizationId, p),
    params,
  );
}

export function useAgingReport() {
  return useOrgListQuery<{ asOf: string; rows: AgingRow[] }, Record<string, never>>(
    "agingReport",
    (organizationId) => getSwantaraService().reports.aging(organizationId),
  );
}

export function useInventoryValuation() {
  return useOrgListQuery<{ rows: InventoryValueRow[] }, Record<string, never>>(
    "inventoryValuation",
    (organizationId) => getSwantaraService().reports.inventoryValuation(organizationId),
  );
}

export function useProfitAndLoss(params?: ListQuery) {
  return useOrgListQuery<{ profitAndLoss: ProfitAndLoss }, ListQuery>(
    "profitAndLoss",
    (organizationId, p) => getSwantaraService().reports.profitAndLoss(organizationId, p),
    params,
  );
}

export function useBalanceSheet() {
  return useOrgListQuery<{ balanceSheet: BalanceSheet }, Record<string, never>>(
    "balanceSheet",
    (organizationId) => getSwantaraService().reports.balanceSheet(organizationId),
  );
}

export function useCashFlowReport() {
  return useOrgListQuery<{ cashFlow: CashFlow }, Record<string, never>>(
    "cashFlowReport",
    (organizationId) => getSwantaraService().reports.cashFlow(organizationId),
  );
}
