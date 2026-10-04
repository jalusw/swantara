import { describe, expect, it } from "vitest";
import type { Account } from "@/lib/services/swantara";
import {
  accountTypeLabels,
  accountTypeTone,
  buildAccountTree,
  countAccounts,
} from "../account-utils";

function account(overrides: Partial<Account>): Account {
  return {
    id: overrides.id ?? 1,
    createdAt: new Date(),
    updatedAt: new Date(),
    organizationId: 1,
    code: overrides.code ?? "1000",
    name: overrides.name ?? "Akun",
    type: overrides.type ?? "asset",
    reconcilable: false,
    currencyCode: null,
    parentId: overrides.parentId ?? null,
    active: overrides.active ?? true,
    ...overrides,
  } as Account;
}

describe("buildAccountTree", () => {
  it("returns empty array for empty input", () => {
    expect(buildAccountTree([])).toEqual([]);
  });

  it("builds single root node", () => {
    const accounts = [account({ id: 1, code: "1000", name: "Kas" })];
    const tree = buildAccountTree(accounts);

    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe(1);
    expect(tree[0]!.children).toEqual([]);
  });

  it("builds nested hierarchy from parent references", () => {
    const accounts = [
      account({ id: 1, code: "1000", name: "Aset" }),
      account({ id: 2, code: "1010", name: "Bank", parentId: 1 }),
      account({ id: 3, code: "1020", name: "Kas", parentId: 1 }),
    ];
    const tree = buildAccountTree(accounts);

    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe(1);
    expect(tree[0]!.children).toHaveLength(2);
    expect(tree[0]!.children[0]!.id).toBe(2);
    expect(tree[0]!.children[1]!.id).toBe(3);
  });

  it("treats missing parent as root", () => {
    const accounts = [account({ id: 1, code: "1010", name: "Child", parentId: 999 })];
    const tree = buildAccountTree(accounts);

    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe(1);
  });

  it("sorts roots and children by code", () => {
    const accounts = [
      account({ id: 3, code: "3000", name: "Ekuitas" }),
      account({ id: 1, code: "1000", name: "Aset" }),
      account({ id: 2, code: "2000", name: "Liability" }),
      account({ id: 4, code: "1020", name: "Kas", parentId: 1 }),
      account({ id: 5, code: "1010", name: "Bank", parentId: 1 }),
    ];
    const tree = buildAccountTree(accounts);

    expect(tree.map((n) => n.code)).toEqual(["1000", "2000", "3000"]);
    expect(tree[0]!.children.map((n) => n.code)).toEqual(["1010", "1020"]);
  });
});

describe("countAccounts", () => {
  it("counts active accounts of given type", () => {
    const accounts = [
      account({ id: 1, type: "asset", active: true }),
      account({ id: 2, type: "asset", active: true }),
      account({ id: 3, type: "asset", active: false }),
      account({ id: 4, type: "liability", active: true }),
    ];

    expect(countAccounts(accounts, "asset")).toBe(2);
  });

  it("returns zero when no accounts match type", () => {
    const accounts = [account({ id: 1, type: "liability" })];

    expect(countAccounts(accounts, "asset")).toBe(0);
  });

  it("returns zero for empty array", () => {
    expect(countAccounts([], "asset")).toBe(0);
  });

  it("counts only active accounts", () => {
    const accounts = [
      account({ id: 1, type: "income", active: false }),
      account({ id: 2, type: "income", active: false }),
    ];

    expect(countAccounts(accounts, "income")).toBe(0);
  });
});

describe("accountTypeTone", () => {
  it("returns correct tone for each type", () => {
    expect(accountTypeTone("asset")).toBe("info");
    expect(accountTypeTone("current_asset")).toBe("info");
    expect(accountTypeTone("fixed_asset")).toBe("info");
    expect(accountTypeTone("liability")).toBe("warning");
    expect(accountTypeTone("equity")).toBe("neutral");
    expect(accountTypeTone("income")).toBe("success");
    expect(accountTypeTone("expense")).toBe("danger");
    expect(accountTypeTone("cogs")).toBe("danger");
    expect(accountTypeTone("receivable")).toBe("info");
    expect(accountTypeTone("payable")).toBe("warning");
    expect(accountTypeTone("bank")).toBe("success");
    expect(accountTypeTone("cash")).toBe("success");
    expect(accountTypeTone("tax")).toBe("neutral");
    expect(accountTypeTone("depreciation")).toBe("warning");
  });
});

describe("accountTypeLabels", () => {
  it("has labels for all account types", () => {
    expect(accountTypeLabels.asset).toBe("Asset");
    expect(accountTypeLabels.liability).toBe("Liability");
    expect(accountTypeLabels.equity).toBe("Ekuitas");
    expect(accountTypeLabels.income).toBe("Income");
    expect(accountTypeLabels.expense).toBe("Biaya");
    expect(accountTypeLabels.receivable).toBe("Receivable");
    expect(accountTypeLabels.payable).toBe("Payable");
    expect(accountTypeLabels.bank).toBe("Bank");
    expect(accountTypeLabels.cash).toBe("Kas");
    expect(accountTypeLabels.cogs).toBe("COGS");
    expect(accountTypeLabels.tax).toBe("Pajak");
    expect(accountTypeLabels.current_asset).toBe("Current asset");
    expect(accountTypeLabels.fixed_asset).toBe("Fixed asset");
    expect(accountTypeLabels.depreciation).toBe("Depreciation");
  });
});
