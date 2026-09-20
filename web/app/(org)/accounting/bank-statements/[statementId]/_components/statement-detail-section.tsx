"use client";

import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { BankStatement, BankStatementLine } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import {
  bankStatementStateLabel,
  bankStatementStateTone,
  statementBalance,
} from "../../_components/statement-utils";

export function StatementDetailSection({ statementId }: { orgId: string; statementId: string }) {
  const statementQuery = useOrgQuery<{ bankStatement: BankStatement }>(
    "bankStatements",
    Number(statementId),
    (organizationId) =>
      getSwantaraService().bankStatements.get(organizationId, Number(statementId)),
  );

  const statement = statementQuery.data?.bankStatement;
  const lines: BankStatementLine[] = [];
  const computedBalance = statement ? statementBalance(statement.balanceStart, lines) : 0;
  const isReconciled = statement != null && Math.abs(computedBalance - statement.balanceEnd) < 0.01;

  if (statementQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading statement..."}</p>;
  }

  if (!statement) {
    return <p className="text-sm text-muted-foreground">{"Bank statement not found."}</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold">{statement.name ?? `BS-${statement.id}`}</h1>
          <p className="text-sm text-muted-foreground">
            {statement.date ? formatDate(statement.date) : "—"}
          </p>
        </div>
        <Badge variant="outline" className={bankStatementStateTone(statement.state)}>
          {bankStatementStateLabel(statement.state)}
        </Badge>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{"Opening balance"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-bold tabular-nums">{formatMoney(statement.balanceStart)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{"Closing balance"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-lg font-bold tabular-nums">{formatMoney(statement.balanceEnd)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{"Computed balance"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p
              className={cn(
                "text-lg font-bold tabular-nums",
                isReconciled ? "text-success" : "text-destructive",
              )}
            >
              {formatMoney(computedBalance)}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{"Statement lines"}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 ">{"Date"}</th>
                  <th className="pb-2 ">{"Ref"}</th>
                  <th className="pb-2 ">{"Narration"}</th>
                  <th className="pb-2 text-right">{"Amount"}</th>
                  <th className="pb-2 ">{"Reconciled"}</th>
                </tr>
              </thead>
              <tbody>
                {lines.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="py-4 text-center text-muted-foreground">
                      {"No statement lines found."}
                    </td>
                  </tr>
                ) : (
                  lines.map((line) => (
                    <tr key={line.id} className="border-b last:border-0">
                      <td className="py-2 text-sm">{line.date ? formatDate(line.date) : "—"}</td>
                      <td className="py-2 text-sm">{line.ref ?? "—"}</td>
                      <td className="py-2 text-sm text-muted-foreground">
                        {line.narration ?? "—"}
                      </td>
                      <td className="py-2 text-right tabular-nums">{formatMoney(line.amount)}</td>
                      <td className="py-2">
                        {line.reconciled ? (
                          <Badge variant="secondary" className="text-xs">
                            {"Reconciled"}
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground text-xs">—</span>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
