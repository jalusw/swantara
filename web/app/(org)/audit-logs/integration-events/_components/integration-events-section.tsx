"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { RotateCcw } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { IntegrationEvent } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

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
  const t = useTranslations("AuditLogs");
  const tCommon = useTranslations("Common");
  const eventStatus = (status: string) =>
    (t as unknown as (k: string) => string)(`eventStatus_${status}`);
  const query = useOrgListQuery<{ integrationEvents: IntegrationEvent[] }, Record<string, never>>(
    "integrationEvents",
    (organizationId) => getSwantaraService().integrationEvents.list(organizationId),
  );

  const integrationEvents = query.data?.integrationEvents ?? [];

  function handleDispatch(id: number) {
    void toast.promise(getSwantaraService().integrationEvents.dispatch(Number(orgId), id), {
      loading: t("dispatching"),
      success: () => {
        void query.refetch();
        return t("eventDispatched");
      },
      error: t("dispatchFailed"),
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
      header: t("colTopic"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.topic}</span>,
    },
    {
      accessorKey: "status",
      header: t("colStatus"),
      cell: ({ row }) => (
        <Badge variant="outline" className={integrationEventStatusTone(row.original.status)}>
          {eventStatus(row.original.status)}
        </Badge>
      ),
    },
    {
      accessorKey: "retries",
      header: t("colRetries"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.retries}</span>,
    },
    {
      accessorKey: "createdAt",
      header: t("colCreatedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      accessorKey: "updatedAt",
      header: t("colUpdatedAt"),
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
            aria-label={tCommon("retry")}
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
          { value: "pending", label: eventStatus("pending") },
          { value: "sent", label: eventStatus("sent") },
          { value: "failed", label: eventStatus("failed") },
        ]}
        searchPlaceholder={t("eventsSearchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allEvents")}
        ariaLabel={t("eventsTitle")}
        emptyTitle={t("eventsEmpty")}
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
