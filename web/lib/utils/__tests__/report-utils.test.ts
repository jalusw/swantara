import { describe, expect, it } from "vitest";
import type {
  AgingRow,
  BalanceSheetAccount,
  InventoryValueRow,
  ProfitAndLossRow,
  TrialBalanceRow,
} from "../../services/swantara";
import {
  agingBucketSortKey,
  agingTotalByType,
  bsGroupBySection,
  groupAgingByType,
  inventoryTotalQuantity,
  inventoryTotalValue,
  plGroupByType,
  trialBalanceIsBalanced,
  trialBalanceTotals,
} from "../report-utils";

const tbRows: TrialBalanceRow[] = [
  {
    accountId: 1,
    code: "1000",
    name: "Cash",
    accountType: "asset",
    openingDebit: 1000,
    openingCredit: 0,
    periodDebit: 500,
    periodCredit: 200,
    closingDebit: 1300,
    closingCredit: 0,
  },
  {
    accountId: 2,
    code: "2000",
    name: "AP",
    accountType: "liability",
    openingDebit: 0,
    openingCredit: 500,
    periodDebit: 0,
    periodCredit: 300,
    closingDebit: 0,
    closingCredit: 800,
  },
];

const agingRows: AgingRow[] = [
  { type: "receivable", currencyCode: "USD", bucket: "current", amount: 1000 },
  { type: "receivable", currencyCode: "USD", bucket: "1-30", amount: 500 },
  { type: "payable", currencyCode: "USD", bucket: "31-60", amount: 300 },
];

const inventoryRows: InventoryValueRow[] = [
  { itemId: 1, productName: "Widget", quantity: 10, value: 100 },
  { itemId: 2, productName: "Gadget", quantity: 5, value: 250 },
];

const plRows: ProfitAndLossRow[] = [
  {
    accountId: 1,
    code: "4000",
    name: "Sales",
    accountType: "revenue",
    amount: 10000,
  },
  {
    accountId: 2,
    code: "5000",
    name: "COGS",
    accountType: "cogs",
    amount: 4000,
  },
  {
    accountId: 3,
    code: "6000",
    name: "Rent",
    accountType: "expenses",
    amount: 1000,
  },
];

const bsAssets: BalanceSheetAccount[] = [
  {
    accountId: 1,
    code: "1000",
    name: "Cash",
    accountType: "current_asset",
    balance: 5000,
  },
];
const bsLiabilities: BalanceSheetAccount[] = [
  {
    accountId: 2,
    code: "2000",
    name: "AP",
    accountType: "current_liability",
    balance: 2000,
  },
];
const bsEquity: BalanceSheetAccount[] = [
  {
    accountId: 3,
    code: "3000",
    name: "Equity",
    accountType: "equity",
    balance: 3000,
  },
];

describe("trialBalanceTotals", () => {
  it("sums debit and credit columns", () => {
    const result = trialBalanceTotals(tbRows);
    expect(result.totalDebit).toBe(1300);
    expect(result.totalCredit).toBe(800);
  });

  it("returns zeros for empty rows", () => {
    expect(trialBalanceTotals([])).toEqual({ totalDebit: 0, totalCredit: 0 });
  });
});

describe("trialBalanceIsBalanced", () => {
  it("returns false when debits != credits", () => {
    expect(trialBalanceIsBalanced(tbRows)).toBe(false);
  });

  it("returns true when balanced", () => {
    const balanced: TrialBalanceRow[] = [
      { ...tbRows[0]!, closingDebit: 100, closingCredit: 0 },
      { ...tbRows[1]!, closingDebit: 0, closingCredit: 100 },
    ];
    expect(trialBalanceIsBalanced(balanced)).toBe(true);
  });
});

describe("agingBucketSortKey", () => {
  it("sorts buckets in correct order", () => {
    expect(agingBucketSortKey("current")).toBe(0);
    expect(agingBucketSortKey("1-30")).toBe(1);
    expect(agingBucketSortKey("31-60")).toBe(2);
    expect(agingBucketSortKey("61-90")).toBe(3);
    expect(agingBucketSortKey("90+")).toBe(4);
  });
});

describe("groupAgingByType", () => {
  it("groups rows by type", () => {
    const grouped = groupAgingByType(agingRows);
    expect(grouped.size).toBe(2);
    expect(grouped.get("receivable")).toHaveLength(2);
    expect(grouped.get("payable")).toHaveLength(1);
  });
});

describe("agingTotalByType", () => {
  it("sums amounts", () => {
    expect(agingTotalByType(agingRows)).toBe(1800);
  });
});

describe("inventoryTotalValue", () => {
  it("sums values", () => {
    expect(inventoryTotalValue(inventoryRows)).toBe(350);
  });
});

describe("inventoryTotalQuantity", () => {
  it("sums quantities", () => {
    expect(inventoryTotalQuantity(inventoryRows)).toBe(15);
  });
});

describe("plGroupByType", () => {
  it("groups rows by account type", () => {
    const grouped = plGroupByType(plRows);
    expect(grouped.size).toBe(3);
    expect(grouped.get("revenue")).toHaveLength(1);
    expect(grouped.get("cogs")).toHaveLength(1);
    expect(grouped.get("expenses")).toHaveLength(1);
  });
});

describe("bsGroupBySection", () => {
  it("creates map with three sections", () => {
    const sections = bsGroupBySection(bsAssets, bsLiabilities, bsEquity);
    expect(sections.size).toBe(3);
    expect(sections.get("assets")).toHaveLength(1);
    expect(sections.get("liabilities")).toHaveLength(1);
    expect(sections.get("equity")).toHaveLength(1);
  });
});
