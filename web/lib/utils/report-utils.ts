import type {
  AgingRow,
  BalanceSheetAccount,
  InventoryValueRow,
  ProfitAndLossRow,
  TrialBalanceRow,
} from "../services/swantara";

export type ReportSection = "revenue" | "cogs" | "expenses" | "assets" | "liabilities" | "equity";

export function trialBalanceTotals(rows: TrialBalanceRow[]) {
  return rows.reduce(
    (acc, row) => ({
      totalDebit: acc.totalDebit + row.closingDebit,
      totalCredit: acc.totalCredit + row.closingCredit,
    }),
    { totalDebit: 0, totalCredit: 0 },
  );
}

export function trialBalanceIsBalanced(rows: TrialBalanceRow[]): boolean {
  const { totalDebit, totalCredit } = trialBalanceTotals(rows);
  return Math.abs(totalDebit - totalCredit) < 0.01;
}

export function agingBucketSortKey(bucket: AgingRow["bucket"]): number {
  const order: Record<AgingRow["bucket"], number> = {
    current: 0,
    "1-30": 1,
    "31-60": 2,
    "61-90": 3,
    "90+": 4,
  };
  return order[bucket];
}

export function groupAgingByType(rows: AgingRow[]): Map<string, AgingRow[]> {
  const map = new Map<string, AgingRow[]>();
  for (const row of rows) {
    const existing = map.get(row.type) ?? [];
    existing.push(row);
    map.set(row.type, existing);
  }
  return map;
}

export function agingTotalByType(rows: AgingRow[]): number {
  return rows.reduce((sum, row) => sum + row.amount, 0);
}

export function inventoryTotalValue(rows: InventoryValueRow[]): number {
  return rows.reduce((sum, row) => sum + row.value, 0);
}

export function inventoryTotalQuantity(rows: InventoryValueRow[]): number {
  return rows.reduce((sum, row) => sum + row.quantity, 0);
}

export function plGroupByType(rows: ProfitAndLossRow[]): Map<string, ProfitAndLossRow[]> {
  const map = new Map<string, ProfitAndLossRow[]>();
  for (const row of rows) {
    const existing = map.get(row.accountType) ?? [];
    existing.push(row);
    map.set(row.accountType, existing);
  }
  return map;
}

export function bsGroupBySection(
  assets: BalanceSheetAccount[],
  liabilities: BalanceSheetAccount[],
  equity: BalanceSheetAccount[],
): Map<ReportSection, BalanceSheetAccount[]> {
  return new Map([
    ["assets", assets],
    ["liabilities", liabilities],
    ["equity", equity],
  ]);
}
