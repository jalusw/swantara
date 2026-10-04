"use client";

import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ExpenseReport } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
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
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Expenses");
  const reportQuery = useOrgListQuery<{ report: ExpenseReport }, Record<string, never>>(
    "expenseReport",
    (organizationId) => getSwantaraService().expenseReports.get(organizationId, Number(expenseId)),
  );

  const report = reportQuery.data?.report;

  if (reportQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{t("loadingReport")}</p>;
  }

  if (!report) {
    return <p className="text-sm text-muted-foreground">{t("reportNotFound")}</p>;
  }

  const tone = expenseStateTone(report.state);

  function handleAction(action: "submit" | "approve" | "refuse" | "post" | "reimburse") {
    void toast.promise(
      getSwantaraService().expenseReports[action](Number(orgId), Number(expenseId)),
      {
        loading: t("processing"),
        success: () => {
          void reportQuery.refetch();
          return t("toastSaved");
        },
        error: t("toastActionFailed"),
      },
    );
  }

  function handleBill() {
    void toast.promise(getSwantaraService().expenseReports.bill(Number(orgId), Number(expenseId)), {
      loading: t("processing"),
      success: () => {
        void reportQuery.refetch();
        return t("toastInvoiceFromBill");
      },
      error: t("toastActionFailed"),
    });
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canSubmit(report.state) ? (
        <Button size="sm" onClick={() => handleAction("submit")}>
          {t("submit")}
        </Button>
      ) : null}
      {canApprove(report.state) ? (
        <Button size="sm" onClick={() => handleAction("approve")}>
          {t("approve")}
        </Button>
      ) : null}
      {canRefuse(report.state) ? (
        <Button size="sm" variant="destructive" onClick={() => handleAction("refuse")}>
          {t("refuse")}
        </Button>
      ) : null}
      {canPost(report.state) ? (
        <Button size="sm" onClick={() => handleAction("post")}>
          {t("post")}
        </Button>
      ) : null}
      {canReimburse(report.state) ? (
        <Button size="sm" onClick={() => handleAction("reimburse")}>
          {t("reimburse")}
        </Button>
      ) : null}
      {canBill(report.state) ? (
        <Button size="sm" onClick={handleBill}>
          {t("bill")}
        </Button>
      ) : null}
    </div>
  );

  const lines = report.lines ?? [];

  return (
    <RecordLayout
      breadcrumbItems={[{ label: t("expensesTitle"), href: "/expenses" }, { label: report.name }]}
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
          {(t as unknown as (k: string) => string)(`expenseState_${report.state}`)}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: t("overviewTab"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("totalAmount")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(report.totalAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("fieldEmployee")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">#{report.employeeId}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("fieldPaymentMode")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <Badge variant="outline">
                    {(t as unknown as (k: string) => string)(`paymentMode_${report.paymentMode}`)}
                  </Badge>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("details")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("submittedAt")}</dt>
                      <dd className="text-sm">
                        {report.submittedAt ? formatDate(String(report.submittedAt)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("approvedBy")}</dt>
                      <dd className="text-sm">
                        {report.approvedBy ? `#${report.approvedBy}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("moveId")}</dt>
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
          label: t("linesTab"),
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t("expenseLinesEmpty")}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{t("colDescription")}</th>
                        <th className="pb-2 text-left">{t("colDate")}</th>
                        <th className="pb-2 text-right">{t("colQuantity")}</th>
                        <th className="pb-2 text-right">{t("colUnitPrice")}</th>
                        <th className="pb-2 text-right">{t("colAmount")}</th>
                        <th className="pb-2 text-right">{t("colReimbursable")}</th>
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
