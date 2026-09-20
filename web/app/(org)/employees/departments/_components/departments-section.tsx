"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, PencilIcon, Plus, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { ConfirmDialog } from "@/components/confirm-dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/dropdown";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Department } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { DepartmentFormDialog } from "./department-form-dialog";

export function DepartmentsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editDepartment, setEditDepartment] = useState<Department | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteDepartment, setDeleteDepartment] = useState<Department | null>(null);

  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const departments = departmentsQuery.data?.departments ?? [];

  function handleSave(_id: string) {
    setDialogOpen(false);
    setEditDepartment(null);
    void departmentsQuery.refetch();
  }

  function handleEdit(department: Department) {
    setEditDepartment(department);
    setDialogOpen(true);
  }

  function handleDeleteClick(department: Department) {
    setDeleteDepartment(department);
    setDeleteDialogOpen(true);
  }

  const columns: ColumnDef<Department>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "description",
      header: "Description",
      cell: ({ row }) =>
        row.original.description ?? <span className="text-muted-foreground">—</span>,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon-sm">
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => handleEdit(row.original)}>
              <PencilIcon />
              {"Edit"}
            </DropdownMenuItem>
            <DropdownMenuItem variant="destructive" onClick={() => handleDeleteClick(row.original)}>
              <Trash2Icon />
              {"Delete"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  const isLoading = departmentsQuery.isLoading;
  const error = departmentsQuery.isError ? departmentsQuery.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={departments}
        getRowId={(row) => String(row.id)}
        searchKeys={["name", "description"]}
        searchPlaceholder={"Search departments…"}
        ariaLabel={"All departments"}
        emptyTitle={"No departments found"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void departmentsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditDepartment(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{"Add department"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <DepartmentFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editDepartment}
          onSave={handleSave}
        />
      ) : null}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title={"Delete department"}
        description={"Are you sure you want to delete this department?"}
        onConfirm={() => {
          if (!deleteDepartment) return;
          void getSwantaraService()
            .departments.delete(Number(orgId), deleteDepartment.id)
            .then(() => {
              setDeleteDepartment(null);
              void departmentsQuery.refetch();
            })
            .catch(() => toast.error("Could not disable the organization."));
        }}
      />
    </div>
  );
}
