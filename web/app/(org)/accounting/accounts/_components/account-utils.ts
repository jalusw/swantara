import type { Account } from "@/lib/services/swantara";

export type AccountType = Account["type"];

export const accountTypeLabels: Record<AccountType, string> = {
  asset: "Asset",
  liability: "Liability",
  equity: "Ekuitas",
  income: "Income",
  expense: "Biaya",
  receivable: "Receivable",
  payable: "Payable",
  bank: "Bank",
  cash: "Kas",
  cogs: "COGS",
  tax: "Pajak",
  current_asset: "Current asset",
  fixed_asset: "Fixed asset",
  depreciation: "Depreciation",
};

export function accountTypeTone(
  type: AccountType,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (type) {
    case "asset":
    case "current_asset":
    case "fixed_asset":
      return "info";
    case "liability":
      return "warning";
    case "equity":
      return "neutral";
    case "income":
      return "success";
    case "expense":
    case "cogs":
      return "danger";
    case "receivable":
      return "info";
    case "payable":
      return "warning";
    case "bank":
    case "cash":
      return "success";
    case "tax":
      return "neutral";
    case "depreciation":
      return "warning";
    default:
      return "neutral";
  }
}

export type AccountTreeNode = Account & { children: AccountTreeNode[] };

export function buildAccountTree(accounts: Account[]): AccountTreeNode[] {
  const map = new Map<number, AccountTreeNode>();
  const roots: AccountTreeNode[] = [];

  for (const account of accounts) {
    map.set(account.id, { ...account, children: [] });
  }

  for (const account of accounts) {
    const node = map.get(account.id)!;
    if (account.parentId != null && map.has(account.parentId)) {
      map.get(account.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  roots.sort((a, b) => a.code.localeCompare(b.code));
  for (const node of map.values()) {
    node.children.sort((a, b) => a.code.localeCompare(b.code));
  }

  return roots;
}

export function countAccounts(accounts: Account[], type: AccountType): number {
  return accounts.filter((a) => a.type === type).reduce((sum, a) => sum + (a.active ? 1 : 0), 0);
}
