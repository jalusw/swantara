"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ExpenseCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { ExpenseCategoryFormDialog } from "./expense-category-form-dialog";

export function ExpenseCategoriesSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCategory, setEditCategory] = useState<ExpenseCategory | null>(null);

  const categoriesQuery = useOrgListQuery<{ categories: ExpenseCategory[] }, Record<string, never>>(
    "expenseCategories",
    (organizationId) => getSwantaraService().expenseCategories.list(organizationId),
  );

  const categories = categoriesQuery.data?.categories ?? [];

  const deleteMutation = useMutation({
    mutationFn: (categoryId: number) =>
      getSwantaraService().expenseCategories.delete(Number(orgId), categoryId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["expenseCategories", Number(orgId)] });
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  function handleSave() {
    setDialogOpen(false);
    setEditCategory(null);
    void categoriesQuery.refetch();
  }

  function handleEdit(category: ExpenseCategory) {
    setEditCategory(category);
    setDialogOpen(true);
  }

  const columns: ColumnDef<ExpenseCategory>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <button
          type="button"
          onClick={() => handleEdit(row.original)}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </button>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Deleted"}
          confirmTitle={"Delete this activity?"}
          confirmDescription={"The activity will be removed permanently."}
          onEdit={() => handleEdit(row.original)}
          onDelete={() => deleteMutation.mutate(row.original.id)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={categories}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search categories…"}
        allLabel={"All categories"}
        ariaLabel={"Expense categories"}
        emptyTitle={"No expense categories yet"}
        status={
          categoriesQuery.isLoading
            ? { type: "loading" }
            : categoriesQuery.isError
              ? {
                  type: "error",
                  message: categoriesQuery.error.message,
                  onRetry: () => void categoriesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditCategory(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{"Create category"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ExpenseCategoryFormDialog
          open={dialogOpen}
          onOpenChange={(open) => {
            setDialogOpen(open);
            if (!open) setEditCategory(null);
          }}
          orgId={orgId}
          category={editCategory}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
