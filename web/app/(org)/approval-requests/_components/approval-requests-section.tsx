"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ApprovalRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function ApprovalRequestsSection() {
  const t = useTranslations("ApprovalRequests");
  const tCommon = useTranslations("Common");
  const requestState = (state: string) => (t as unknown as (k: string) => string)(`state_${state}`);
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
      header: t("colType"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.ownerType}</span>,
    },
    {
      accessorKey: "ownerId",
      header: t("colRecord"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{`#${row.original.ownerId}`}</span>
      ),
    },
    {
      accessorKey: "requestedBy",
      header: t("colRequestedBy"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{`#${row.original.requestedBy}`}</span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => (
        <StateBadge
          value={row.original.state}
          statuses={{
            pending: { label: requestState("pending"), tone: "warning" },
            approved: { label: requestState("approved"), tone: "success" },
            refused: { label: requestState("refused"), tone: "danger" },
          }}
        />
      ),
    },
    {
      accessorKey: "createdAt",
      header: t("colCreatedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("view")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteTitle")}
          confirmDescription={t("deleteDescription")}
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
          { value: "pending", label: requestState("pending") },
          { value: "approved", label: requestState("approved") },
          { value: "refused", label: requestState("refused") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allRequests")}
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
