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
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { DeferralSchedule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { DeferralFormDialog } from "./deferral-form-dialog";
import { deferralStateTone } from "./deferral-utils";

export function DeferralsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Deferrals");
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const schedulesQuery = useOrgListQuery<{ schedules: DeferralSchedule[] }, Record<string, never>>(
    "deferrals",
    (organizationId) => getSwantaraService().deferrals.list(organizationId),
  );

  const schedules = schedulesQuery.data?.schedules ?? [];

  function handleSave(_id: string) {
    setDialogOpen(false);
    void schedulesQuery.refetch();
  }

  const columns: ColumnDef<DeferralSchedule>[] = [
    {
      accessorKey: "sourceType",
      header: () => t("colSource"),
      cell: ({ row }) => (
        <a
          href={`/deferrals/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.sourceType} #{row.original.sourceId}
        </a>
      ),
    },
    {
      accessorKey: "type",
      header: () => t("colType"),
      cell: ({ row }) => (
        <Badge variant="outline">
          {(t as unknown as (k: string) => string)(`deferralType_${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "totalAmount",
      header: () => t("colTotal"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.totalAmount)}</span>
      ),
    },
    {
      accessorKey: "recognizedAmount",
      header: () => t("colRecognized"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatNumber(row.original.recognizedAmount)}</span>
      ),
    },
    {
      accessorKey: "method",
      header: () => t("colMethod"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {(t as unknown as (k: string) => string)(`deferMethod_${row.original.method}`)}
        </span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: () => t("colStartDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateStart ? formatDate(String(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: () => t("colStatus"),
      cell: ({ row }) => {
        const tone = deferralStateTone(row.original.state);
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
            {(t as unknown as (k: string) => string)(`deferralState_${row.original.state}`)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("actionView")}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/deferrals/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={schedules}
        getRowId={(row) => String(row.id)}
        searchKeys={["sourceType"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: (t as unknown as (k: string) => string)("deferralState_draft") },
          {
            value: "running",
            label: (t as unknown as (k: string) => string)("deferralState_running"),
          },
          { value: "done", label: (t as unknown as (k: string) => string)("deferralState_done") },
          {
            value: "cancelled",
            label: (t as unknown as (k: string) => string)("deferralState_cancelled"),
          },
        ]}
        searchPlaceholder={t("searchDeferrals")}
        filterLabel={t("colStatus")}
        allLabel={t("allDeferrals")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyDeferrals")}
        status={
          schedulesQuery.isLoading
            ? { type: "loading" }
            : schedulesQuery.isError
              ? {
                  type: "error",
                  message: schedulesQuery.error.message,
                  onRetry: () => void schedulesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("createDeferral")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <DeferralFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
