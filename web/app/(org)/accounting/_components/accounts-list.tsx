"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { formatMoney } from "@/lib/utils";
import type { getMockAccounts } from "../_utils";

export function AccountsList({ accounts }: { accounts: ReturnType<typeof getMockAccounts> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Accounts"}</CardTitle>
        <CardDescription>{"Current balances by account group."}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {accounts.map((account) => (
            <div
              key={account.name}
              className="flex items-center justify-between rounded-lg border border-border p-3"
            >
              <span className="text-sm">{account.name}</span>
              <div className="flex items-baseline gap-2">
                <span className={account.movement === "up" ? "text-success" : "text-destructive"}>
                  {account.percent}
                </span>
                <span className="text-sm tabular-nums">
                  {formatMoney(account.balance, { currency: DEFAULT_CURRENCY })}
                </span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
