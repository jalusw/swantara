"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Department, JobPosition } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { JobPositionFormDialog } from "./job-position-form-dialog";

export function JobPositionsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<JobPosition | null>(null);

  const query = useOrgListQuery<{ jobPositions: JobPosition[] }, Record<string, never>>(
    "jobPositions",
    (organizationId) => getSwantaraService().jobPositions.list(organizationId),
  );

  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const jobPositions = query.data?.jobPositions ?? [];
  const departments = departmentsQuery.data?.departments ?? [];

  function handleOpenCreate() {
    setEditing(null);
    setDialogOpen(true);
  }

  function handleOpenEdit(position: JobPosition) {
    setEditing(position);
    setDialogOpen(true);
  }

  function handleSave() {
    setDialogOpen(false);
    setEditing(null);
    void query.refetch();
  }

  const columns: ColumnDef<JobPosition>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "departmentId",
      header: "Department",
      cell: ({ row }) => (
        <Badge variant="secondary">
          {departments.find((dept) => dept.id === row.original.departmentId)?.name ?? "—"}
        </Badge>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete position"}
          confirmDescription={"Are you sure you want to delete this position?"}
          onEdit={() => handleOpenEdit(row.original)}
          onDelete={() =>
            void getSwantaraService()
              .jobPositions.delete(Number(orgId), row.original.id)
              .then(() => void query.refetch())
              .catch(() => toast.error("Something went wrong."))
          }
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={jobPositions}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search positions…"}
        filterLabel=""
        statusOptions={[]}
        allLabel=""
        ariaLabel={"Job positions"}
        emptyTitle={"No positions found"}
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
          <Button size="sm" onClick={handleOpenCreate}>
            <Plus />
            <span>{"Add position"}</span>
          </Button>
        }
      />
      <JobPositionFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        orgId={orgId}
        initial={editing}
        onSave={handleSave}
      />
    </div>
  );
}
