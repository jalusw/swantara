"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useRouter } from "next/navigation";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ApprovalRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function ApprovalRequestsSection() {
  const router = useRouter();

  const query = useOrgListQuery<{ approvalRequests: ApprovalRequest[] }, Record<string, never>>(
    "approvalRequests",
    (organizationId) => getSwantaraService().approvalRequests.list(organizationId),
  );

  const approvalRequests = query.data?.approvalRequests ?? [];

  const columns: ColumnDef<ApprovalRequest>[] = [
    {
      accessorKey: "id",
      header: "ID",
      cell: ({ row }) => (
        <a
          href={`/approval-requests/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {`AR-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "ownerType",
      header: "Type",
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.ownerType}</span>,
    },
    {
      accessorKey: "ownerId",
      header: "Record",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{`#${row.original.ownerId}`}</span>
      ),
    },
    {
      accessorKey: "requestedBy",
      header: "Requested by",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{`#${row.original.requestedBy}`}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <StateBadge
          value={row.original.state}
          statuses={{
            pending: { label: "Pending", tone: "warning" },
            approved: { label: "Approved", tone: "success" },
            refused: { label: "Refused", tone: "danger" },
          }}
        />
      ),
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
          deleteLabel={"Deleted"}
          confirmTitle={"Delete this activity?"}
          confirmDescription={"The activity will be removed permanently."}
          onEdit={() => router.push(`/approval-requests/${row.original.id}`)}
        />
      ),
    },
  ];

  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={approvalRequests}
        getRowId={(row) => String(row.id)}
        searchKeys={["ownerType"]}
        statusKey="state"
        statusOptions={[
          { value: "pending", label: "Pending" },
          { value: "approved", label: "Approved" },
          { value: "refused", label: "Refused" },
        ]}
        searchPlaceholder={"Search approvals…"}
        filterLabel={"State"}
        allLabel={"All requests"}
        ariaLabel={"Approval requests"}
        emptyTitle={"No approval requests"}
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
