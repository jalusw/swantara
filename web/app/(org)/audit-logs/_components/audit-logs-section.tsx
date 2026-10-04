"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { AuditLog } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

function auditLogActionTone(action: AuditLog["action"]): string {
  switch (action) {
    case "insert":
      return "border-success text-success";
    case "delete":
      return "border-destructive text-destructive";
    default:
      return "border-warning text-warning";
  }
}

export function AuditLogsSection({ orgId: _orgId }: { orgId: string }) {
  const t = useTranslations("AuditLogs");
  const actionLabel = (action: string) =>
    (t as unknown as (k: string) => string)(`action_${action}`);
  const query = useOrgListQuery<{ auditLogs: AuditLog[] }, Record<string, never>>(
    "auditLogs",
    (organizationId) => getSwantaraService().auditLogs.list(organizationId),
  );

  const auditLogs = query.data?.auditLogs ?? [];

  const columns: ColumnDef<AuditLog>[] = [
    {
      accessorKey: "id",
      header: "ID",
      cell: ({ row }) => <span className="">{`AL-${row.original.id}`}</span>,
    },
    {
      accessorKey: "tableName",
      header: t("colTable"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.tableName}</span>,
    },
    {
      accessorKey: "recordId",
      header: t("colRecord"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.recordId}</span>,
    },
    {
      accessorKey: "action",
      header: t("colAction"),
      cell: ({ row }) => (
        <Badge variant="outline" className={auditLogActionTone(row.original.action)}>
          {actionLabel(row.original.action)}
        </Badge>
      ),
    },
    {
      accessorKey: "changedBy",
      header: t("colChangedBy"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.changedBy}</span>,
    },
    {
      accessorKey: "changedAt",
      header: t("colChangedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.changedAt)}</span>
      ),
    },
    {
      accessorKey: "createdAt",
      header: t("colCreatedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
  ];

  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={auditLogs}
        getRowId={(row) => String(row.id)}
        searchKeys={["tableName"]}
        statusKey="action"
        statusOptions={[
          { value: "insert", label: actionLabel("insert") },
          { value: "update", label: actionLabel("update") },
          { value: "delete", label: actionLabel("delete") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("colAction")}
        allLabel={t("allActions")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyTitle")}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
      />
    </div>
  );
}
