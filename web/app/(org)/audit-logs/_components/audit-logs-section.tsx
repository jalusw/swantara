"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { AuditLog } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

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

function auditLogActionLabel(action: AuditLog["action"]): string {
  switch (action) {
    case "insert":
      return "Created";
    case "update":
      return "Updated";
    case "delete":
      return "Deleted";
    default:
      return humanizeKey(String(action));
  }
}

export function AuditLogsSection({ orgId: _orgId }: { orgId: string }) {
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
      header: "Table",
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.tableName}</span>,
    },
    {
      accessorKey: "recordId",
      header: "Record",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.recordId}</span>,
    },
    {
      accessorKey: "action",
      header: "Action",
      cell: ({ row }) => (
        <Badge variant="outline" className={auditLogActionTone(row.original.action)}>
          {auditLogActionLabel(row.original.action)}
        </Badge>
      ),
    },
    {
      accessorKey: "changedBy",
      header: "Changed by",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.changedBy}</span>,
    },
    {
      accessorKey: "changedAt",
      header: "Changed at",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.changedAt)}</span>
      ),
    },
    {
      accessorKey: "createdAt",
      header: "Created",
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
          { value: "insert", label: "Created" },
          { value: "update", label: "Updated" },
          { value: "delete", label: "Deleted" },
        ]}
        searchPlaceholder={"Search audit logs…"}
        filterLabel={"Action"}
        allLabel={"All actions"}
        ariaLabel={"Audit logs"}
        emptyTitle={"No audit logs"}
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
