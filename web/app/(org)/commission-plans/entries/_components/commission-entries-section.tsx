"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CommissionEntry } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function CommissionEntriesSection({ orgId: _orgId }: { orgId: string }) {
  const query = useOrgListQuery<{ commissionEntries: CommissionEntry[] }, Record<string, never>>(
    "commissionEntries",
    (organizationId) => getSwantaraService().commissionEntries.list(organizationId),
  );

  const entries = query.data?.commissionEntries ?? [];

  const totalCommission = entries.reduce((sum, e) => sum + e.commissionAmount, 0);

  const columns: ColumnDef<CommissionEntry>[] = [
    {
      accessorKey: "salespersonId",
      header: "Salesperson",
      cell: ({ row }) => <span className="text-sm">#{row.original.salespersonId}</span>,
    },
    {
      accessorKey: "planId",
      header: "Plan",
      cell: ({ row }) => <span className="text-sm">#{row.original.planId}</span>,
    },
    {
      accessorKey: "sourceType",
      header: "Source type",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.sourceType}</span>
      ),
    },
    {
      accessorKey: "baseAmount",
      header: "Base",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.baseAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "commissionAmount",
      header: "Commission",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.commissionAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
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
            {humanizeKey(String(row.original.state))}
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
            <CardTitle>{"Total Entries"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{entries.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{"Total Commission"}</CardTitle>
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
          { value: "draft", label: "Draft" },
          { value: "confirmed", label: "Confirmed" },
          { value: "paid", label: "Paid" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search entries…"}
        filterLabel={"State"}
        allLabel={"All entries"}
        ariaLabel={"Commission entries"}
        emptyTitle={"No commission entries yet"}
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
