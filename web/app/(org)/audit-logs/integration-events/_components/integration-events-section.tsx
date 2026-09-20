"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { RotateCcw } from "lucide-react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { IntegrationEvent } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

function integrationEventStatusTone(status: IntegrationEvent["status"]): string {
  switch (status) {
    case "sent":
      return "border-success text-success";
    case "failed":
      return "border-destructive text-destructive";
    default:
      return "";
  }
}

export function IntegrationEventsSection({ orgId }: { orgId: string }) {
  const query = useOrgListQuery<{ integrationEvents: IntegrationEvent[] }, Record<string, never>>(
    "integrationEvents",
    (organizationId) => getSwantaraService().integrationEvents.list(organizationId),
  );

  const integrationEvents = query.data?.integrationEvents ?? [];

  function handleDispatch(id: number) {
    void toast.promise(getSwantaraService().integrationEvents.dispatch(Number(orgId), id), {
      loading: "Dispatching…",
      success: () => {
        void query.refetch();
        return "Event dispatched";
      },
      error: "Failed to dispatch",
    });
  }

  const columns: ColumnDef<IntegrationEvent>[] = [
    {
      accessorKey: "id",
      header: "ID",
      cell: ({ row }) => <span className="">{`IE-${row.original.id}`}</span>,
    },
    {
      accessorKey: "topic",
      header: "Topic",
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.topic}</span>,
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <Badge variant="outline" className={integrationEventStatusTone(row.original.status)}>
          {humanizeKey(String(row.original.status))}
        </Badge>
      ),
    },
    {
      accessorKey: "retries",
      header: "Retries",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.retries}</span>,
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      accessorKey: "updatedAt",
      header: "Updated",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.updatedAt)}</span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        row.original.status === "failed" ? (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => handleDispatch(row.original.id)}
            aria-label={"Retry"}
          >
            <RotateCcw className="h-4 w-4" />
          </Button>
        ) : null,
    },
  ];

  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={integrationEvents}
        getRowId={(row) => String(row.id)}
        searchKeys={["topic"]}
        statusKey="status"
        statusOptions={[
          { value: "pending", label: "Pending" },
          { value: "sent", label: "Sent" },
          { value: "failed", label: "Failed" },
        ]}
        searchPlaceholder={"Search events…"}
        filterLabel={"Status"}
        allLabel={"All events"}
        ariaLabel={"Integration events"}
        emptyTitle={"No integration events"}
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
