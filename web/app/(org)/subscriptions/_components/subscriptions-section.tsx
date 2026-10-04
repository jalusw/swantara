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
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Subscription, SubscriptionMetrics } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { SubscriptionFormDialog } from "./subscription-form-dialog";
import { subscriptionStateTone } from "./subscription-utils";

export function SubscriptionsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Subscriptions");
  const subState = (state: string) => (t as unknown as (k: string) => string)(`state_${state}`);
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const subsQuery = useOrgListQuery<{ subscriptions: Subscription[] }, Record<string, never>>(
    "subscriptions",
    (organizationId) => getSwantaraService().subscriptions.list(organizationId),
  );

  const metricsQuery = useOrgListQuery<{ metrics: SubscriptionMetrics }, Record<string, never>>(
    "subscriptionMetrics",
    (organizationId) => getSwantaraService().subscriptions.metrics(organizationId),
  );

  const subs = subsQuery.data?.subscriptions ?? [];
  const metrics = metricsQuery.data?.metrics;

  function handleSave(subId: string) {
    setDialogOpen(false);
    void subsQuery.refetch();
    router.push(`/subscriptions/${subId}`);
  }

  const columns: ColumnDef<Subscription>[] = [
    {
      accessorKey: "name",
      header: t("colName"),
      cell: ({ row }) => (
        <a
          href={`/subscriptions/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "mrr",
      header: "MRR",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.mrr, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: t("colStartDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateStart ? formatDate(String(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "nextInvoiceDate",
      header: t("colNextInvoice"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.nextInvoiceDate ? formatDate(String(row.original.nextInvoiceDate)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => {
        const tone = subscriptionStateTone(row.original.state);
        return (
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
            {subState(row.original.state)}
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
          onEdit={() => router.push(`/subscriptions/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      {metrics ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardHeader>
              <CardTitle>{"MRR"}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatMoney(metrics.mrr, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{"ARR"}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatMoney(metrics.arr, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("churned")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">{metrics.churned}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("churnRate")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {(metrics.churnRate * 100).toFixed(1)}%
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{"LTV"}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold tabular-nums">
                {formatMoney(metrics.ltv, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
        </div>
      ) : null}
      <InteractiveEntityTable
        columns={columns}
        data={subs}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: subState("draft") },
          { value: "active", label: subState("active") },
          { value: "paused", label: subState("paused") },
          { value: "churned", label: subState("churned") },
          { value: "closed", label: subState("closed") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allSubscriptions")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyTitle")}
        status={
          subsQuery.isLoading
            ? { type: "loading" }
            : subsQuery.isError
              ? {
                  type: "error",
                  message: subsQuery.error.message,
                  onRetry: () => void subsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newSubscription")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <SubscriptionFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
