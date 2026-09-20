"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { ActiveBadge } from "@/components/active-badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

import { LeaveTypeFormDialog } from "./leave-type-form-dialog";

export function LeaveTypesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<LeaveType | null>(null);

  const query = useOrgListQuery<{ leaveTypes: LeaveType[] }, Record<string, never>>(
    "leaveTypes",
    (organizationId) => getSwantaraService().leaveTypes.list(organizationId),
  );

  const leaveTypes = query.data?.leaveTypes ?? [];

  function openCreate() {
    setEditing(null);
    setDialogOpen(true);
  }

  function openEdit(leaveType: LeaveType) {
    setEditing(leaveType);
    setDialogOpen(true);
  }

  const columns: ColumnDef<LeaveType>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "paid",
      header: "Paid",
      cell: ({ row }) => (
        <ActiveBadge active={row.original.paid}>
          {row.original.paid ? "Paid" : "Unpaid"}
        </ActiveBadge>
      ),
    },
    {
      accessorKey: "allocationDays",
      header: "Allocation days",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {row.original.allocationDays ?? "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this leave type?"}
          confirmDescription={"The leave type will no longer be available."}
          onEdit={() => openEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .leaveTypes.delete(Number(orgId), row.original.id)
              .then(() => void query.refetch())
              .catch(() => toast.error("Could not disable the organization."))
          }
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={leaveTypes}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search leave types…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"Leave types"}
        emptyTitle={"No leave types"}
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
          <Button size="sm" onClick={openCreate}>
            <Plus />
            <span>{"Add leave type"}</span>
          </Button>
        }
      />

      <LeaveTypeFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        orgId={orgId}
        initial={editing}
        onSave={() => {
          setDialogOpen(false);
          setEditing(null);
          void query.refetch();
        }}
      />
    </div>
  );
}
