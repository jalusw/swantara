"use client";

import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ExpenseReport } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import {
  canApprove,
  canBill,
  canPost,
  canRefuse,
  canReimburse,
  canSubmit,
  expenseStateTone,
} from "../../_components/expense-utils";

export function ExpenseDetail({ orgId, expenseId }: { orgId: string; expenseId: string }) {
  const reportQuery = useOrgListQuery<{ report: ExpenseReport }, Record<string, never>>(
    "expenseReport",
    (organizationId) => getSwantaraService().expenseReports.get(organizationId, Number(expenseId)),
  );

  const report = reportQuery.data?.report;

  if (reportQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!report) {
    return <p className="text-sm text-muted-foreground">{"Expense report not found."}</p>;
  }

  const tone = expenseStateTone(report.state);

  function handleAction(action: "submit" | "approve" | "refuse" | "post" | "reimburse") {
    void toast.promise(
      getSwantaraService().expenseReports[action](Number(orgId), Number(expenseId)),
      {
        loading: "Processing…",
        success: () => {
          void reportQuery.refetch();
          return "Saved.";
        },
        error: "Action failed",
      },
    );
  }

  function handleBill() {
    void toast.promise(getSwantaraService().expenseReports.bill(Number(orgId), Number(expenseId)), {
      loading: "Processing…",
      success: () => {
        void reportQuery.refetch();
        return "Invoice created from billable lines";
      },
      error: "Action failed",
    });
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canSubmit(report.state) ? (
        <Button size="sm" onClick={() => handleAction("submit")}>
          {"Submit"}
        </Button>
      ) : null}
      {canApprove(report.state) ? (
        <Button size="sm" onClick={() => handleAction("approve")}>
          {"Approve"}
        </Button>
      ) : null}
      {canRefuse(report.state) ? (
        <Button size="sm" variant="destructive" onClick={() => handleAction("refuse")}>
          {"Refuse"}
        </Button>
      ) : null}
      {canPost(report.state) ? (
        <Button size="sm" onClick={() => handleAction("post")}>
          {"Post"}
        </Button>
      ) : null}
      {canReimburse(report.state) ? (
        <Button size="sm" onClick={() => handleAction("reimburse")}>
          {"Reimburse"}
        </Button>
      ) : null}
      {canBill(report.state) ? (
        <Button size="sm" onClick={handleBill}>
          {"Bill"}
        </Button>
      ) : null}
    </div>
  );

  const lines = report.lines ?? [];

  return (
    <RecordLayout
      breadcrumbItems={[{ label: "Expenses", href: "/expenses" }, { label: report.name }]}
      title={report.name}
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
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : ""
          }
        >
          {humanizeKey(String(report.state))}
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
                  <CardTitle>{"Total amount"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(report.totalAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Employee"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">#{report.employeeId}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Payment mode"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <Badge variant="outline">{humanizeKey(String(report.paymentMode))}</Badge>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{"Details"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Submitted at"}</dt>
                      <dd className="text-sm">
                        {report.submittedAt ? formatDate(String(report.submittedAt)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Approved by"}</dt>
                      <dd className="text-sm">
                        {report.approvedBy ? `#${report.approvedBy}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Move Id"}</dt>
                      <dd className="text-sm">{report.entryId ? `#${report.entryId}` : "—"}</dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "lines",
          label: "Lines",
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No expense lines."}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"Description"}</th>
                        <th className="pb-2 text-left">{"Date"}</th>
                        <th className="pb-2 text-right">{"Quantity"}</th>
                        <th className="pb-2 text-right">{"Unit price"}</th>
                        <th className="pb-2 text-right">{"Amount"}</th>
                        <th className="pb-2 text-right">{"Reimbursable"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => (
                        <tr key={line.id} className="border-b">
                          <td className="py-2">{line.description ?? "—"}</td>
                          <td className="py-2 text-muted-foreground">
                            {line.expenseDate ? formatDate(String(line.expenseDate)) : "—"}
                          </td>
                          <td className="py-2 text-right tabular-nums">{line.quantity}</td>
                          <td className="py-2 text-right tabular-nums">
                            {formatNumber(line.unitPrice)}
                          </td>
                          <td className="py-2 text-right tabular-nums ">
                            {formatNumber(line.amount)}
                          </td>
                          <td className="py-2 text-right">{line.reimbursable ? "✓" : "—"}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ),
        },
      ]}
    />
  );
}
