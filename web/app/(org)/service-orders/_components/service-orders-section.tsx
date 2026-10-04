"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ServiceOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { ServiceOrderFormDialog } from "./service-order-form-dialog";
import { serviceOrderStateTone } from "./service-order-utils";

export function ServiceOrdersSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Service");
  const orderState = (state: string) =>
    (t as unknown as (k: string) => string)(`orderState_${state}`);
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ serviceOrders: ServiceOrder[] }, Record<string, never>>(
    "serviceOrders",
    (organizationId) => getSwantaraService().serviceOrders.list(organizationId),
  );

  const orders = query.data?.serviceOrders ?? [];

  function handleSave(id: string) {
    setDialogOpen(false);
    void query.refetch();
    router.push(`/service-orders/${id}`);
  }

  const columns: ColumnDef<ServiceOrder>[] = [
    {
      accessorKey: "name",
      header: t("colName"),
      cell: ({ row }) => (
        <a
          href={`/service-orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "type",
      header: t("colType"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{String(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "priority",
      header: t("colPriority"),
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.priority}</span>,
    },
    {
      accessorKey: "scheduledDate",
      header: t("colScheduled"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.scheduledDate ? formatDate(String(row.original.scheduledDate)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => {
        const tone = serviceOrderStateTone(row.original.state);
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : tone === "info"
                      ? "border-info text-info"
                      : ""
            }
          >
            {orderState(row.original.state)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("view")}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/service-orders/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader>
            <CardTitle>{t("totalOrders")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{orders.length}</p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={orders}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "new", label: orderState("new") },
          { value: "scheduled", label: orderState("scheduled") },
          { value: "in_progress", label: orderState("in_progress") },
          { value: "done", label: orderState("done") },
          { value: "invoiced", label: orderState("invoiced") },
          { value: "cancelled", label: orderState("cancelled") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allOrders")}
        ariaLabel={t("ordersTitle")}
        emptyTitle={t("emptyTitle")}
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
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newOrder")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ServiceOrderFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
