"use client";

import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { PayrollRun, Payslip } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber, getLocalDateString } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import {
  canClose,
  canConfirm,
  canPay,
  payrollRunStateTone,
  totalGross,
  totalNet,
} from "../../_components/payroll-utils";

export function PayrollRunDetail({ orgId, runId }: { orgId: string; runId: string }) {
  const runQuery = useOrgListQuery<{ run: PayrollRun; payslips: Payslip[] }, Record<string, never>>(
    "payrollRun",
    (organizationId) => getSwantaraService().payrollRuns.get(organizationId, Number(runId)),
  );

  const run = runQuery.data?.run;
  const payslips = runQuery.data?.payslips ?? [];

  if (runQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!run) {
    return <p className="text-sm text-muted-foreground">{"Payroll run not found."}</p>;
  }

  const tone = payrollRunStateTone(run.state);

  function handleAction(action: "confirm" | "pay" | "close") {
    const service = getSwantaraService().payrollRuns;
    const orgNum = Number(orgId);
    const runNum = Number(runId);

    if (action === "confirm") {
      void toast.promise(
        service.confirm(orgNum, runNum, {
          journalId: 1,
          date: getLocalDateString(),
        }),
        {
          loading: "Processing…",
          success: () => {
            void runQuery.refetch();
            return "Payroll run confirmed";
          },
          error: "Action failed",
        },
      );
    } else if (action === "pay") {
      void toast.promise(
        service.pay(orgNum, runNum, {
          journalId: 1,
          date: getLocalDateString(),
        }),
        {
          loading: "Processing…",
          success: () => {
            void runQuery.refetch();
            return "Payroll run paid";
          },
          error: "Action failed",
        },
      );
    } else {
      void toast.promise(service.close(orgNum, runNum), {
        loading: "Processing…",
        success: () => {
          void runQuery.refetch();
          return "Payroll run closed";
        },
        error: "Action failed",
      });
    }
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canConfirm(run.state) ? (
        <Button size="sm" onClick={() => handleAction("confirm")}>
          {"Confirm"}
        </Button>
      ) : null}
      {canPay(run.state) ? (
        <Button size="sm" onClick={() => handleAction("pay")}>
          {"Pay"}
        </Button>
      ) : null}
      {canClose(run.state) ? (
        <Button size="sm" variant="outline" onClick={() => handleAction("close")}>
          {"Close"}
        </Button>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Payroll runs", href: "/payroll-runs" },
        { label: run.name ?? `PR-${run.id}` },
      ]}
      title={run.name ?? `PR-${run.id}`}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "info"
                  ? "border-info text-info"
                  : ""
          }
        >
          {humanizeKey(String(run.state))}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Period"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {formatDate(run.periodStart)} – {formatDate(run.periodEnd)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Gross total"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(totalGross(payslips))}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Net total"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(totalNet(payslips))}
                  </p>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "payslips",
          label: "Payslips",
          content: (
            <div className="space-y-3">
              {payslips.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No payslips computed yet"}</p>
              ) : (
                payslips.map((payslip) => <PayslipCard key={payslip.id} payslip={payslip} />)
              )}
            </div>
          ),
        },
      ]}
    />
  );
}

function PayslipCard({ payslip }: { payslip: Payslip }) {
  return (
    <Card>
      <CardContent>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm">
              {"Employee"} #{payslip.employeeId}
            </p>
            <p className="text-xs text-muted-foreground">
              {"Contract"} #{payslip.contractId}
            </p>
          </div>
          <div className="text-right">
            <p className="text-sm tabular-nums">
              {"Gross"}: {formatNumber(payslip.gross)}
            </p>
            <p className="text-sm tabular-nums">
              {"Net"}: {formatNumber(payslip.net)}
            </p>
          </div>
        </div>
        {payslip.lines.length > 0 ? (
          <div className="mt-3 border-t pt-3">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-muted-foreground">
                  <th className="text-left">{"Rule"}</th>
                  <th className="text-left">{"Category"}</th>
                  <th className="text-right">{"Amount"}</th>
                </tr>
              </thead>
              <tbody>
                {payslip.lines.map((line) => (
                  <tr key={line.id}>
                    <td>{line.name}</td>
                    <td>
                      <Badge variant={line.category === "earning" ? "default" : "destructive"}>
                        {line.category}
                      </Badge>
                    </td>
                    <td className="text-right tabular-nums">{formatNumber(line.amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
