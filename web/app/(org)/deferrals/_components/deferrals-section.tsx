"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { DeferralSchedule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { DeferralFormDialog } from "./deferral-form-dialog";
import { deferralStateTone } from "./deferral-utils";

export function DeferralsSection({ orgId }: { orgId: string }) {
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
      header: "Source",
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
      header: "Type",
      cell: ({ row }) => <Badge variant="outline">{String(row.original.type)}</Badge>,
    },
    {
      accessorKey: "totalAmount",
      header: "Total",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.totalAmount)}</span>
      ),
    },
    {
      accessorKey: "recognizedAmount",
      header: "Recognized",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatNumber(row.original.recognizedAmount)}</span>
      ),
    },
    {
      accessorKey: "method",
      header: "Method",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{String(row.original.method)}</span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: "Start date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateStart ? formatDate(String(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
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
            {humanizeKey(String(row.original.state))}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
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
          { value: "draft", label: "Draft" },
          { value: "running", label: "Running" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search deferrals…"}
        filterLabel={"State"}
        allLabel={"All deferrals"}
        ariaLabel={"Deferrals"}
        emptyTitle={"No deferral schedules yet"}
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
            <span>{"Create deferral"}</span>
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
