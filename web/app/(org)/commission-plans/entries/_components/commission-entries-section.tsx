"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CommissionEntry } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";

export function CommissionEntriesSection({ orgId: _orgId }: { orgId: string }) {
  const t = useTranslations("Commissions");
  const query = useOrgListQuery<{ commissionEntries: CommissionEntry[] }, Record<string, never>>(
    "commissionEntries",
    (organizationId) => getSwantaraService().commissionEntries.list(organizationId),
  );

  const entries = query.data?.commissionEntries ?? [];

  const totalCommission = entries.reduce((sum, e) => sum + e.commissionAmount, 0);

  function stateLabel(value: string): string {
    return (t as unknown as (k: string) => string)(`state_${value}`);
  }

  const columns: ColumnDef<CommissionEntry>[] = [
    {
      accessorKey: "salespersonId",
      header: t("tableSalesperson"),
      cell: ({ row }) => <span className="text-sm">#{row.original.salespersonId}</span>,
    },
    {
      accessorKey: "planId",
      header: t("tablePlan"),
      cell: ({ row }) => <span className="text-sm">#{row.original.planId}</span>,
    },
    {
      accessorKey: "sourceType",
      header: t("tableSourceType"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.sourceType}</span>
      ),
    },
    {
      accessorKey: "baseAmount",
      header: t("tableBase"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.baseAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "commissionAmount",
      header: t("tableCommission"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.commissionAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const tone =
          row.original.state === "confirmed"
            ? "success"
            : row.original.state === "paid"
              ? "info"
              : row.original.state === "cancelled"
                ? "danger"
                : "";
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "info"
                  ? "border-info text-info"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : ""
            }
          >
            {stateLabel(String(row.original.state))}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{t("totalEntries")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{entries.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("totalCommission")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">
              {formatMoney(totalCommission, { currency: DEFAULT_CURRENCY })}
            </p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={entries}
        getRowId={(row) => String(row.id)}
        searchKeys={["sourceType"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: stateLabel("draft") },
          { value: "confirmed", label: stateLabel("confirmed") },
          { value: "paid", label: stateLabel("paid") },
          { value: "cancelled", label: stateLabel("cancelled") },
        ]}
        searchPlaceholder={t("searchEntriesPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allEntries")}
        ariaLabel={t("entriesLabel")}
        emptyTitle={t("emptyEntries")}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
      />
    </div>
  );
}
