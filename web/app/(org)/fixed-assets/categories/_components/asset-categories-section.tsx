"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { AssetCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { AssetCategoryFormDialog } from "./asset-category-form-dialog";

export function AssetCategoriesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCategory, setEditCategory] = useState<AssetCategory | null>(null);

  const categoriesQuery = useOrgListQuery<
    { assetCategories: AssetCategory[] },
    Record<string, never>
  >("assetCategories", (organizationId) =>
    getSwantaraService().assetCategories.list(organizationId),
  );

  const categories = categoriesQuery.data?.assetCategories ?? [];

  function handleSave(_id: string) {
    setDialogOpen(false);
    setEditCategory(null);
    void categoriesQuery.refetch();
  }

  function handleEdit(category: AssetCategory) {
    setEditCategory(category);
    setDialogOpen(true);
  }

  const columns: ColumnDef<AssetCategory>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "method",
      header: "Depreciation method",
      cell: ({ row }) => <Badge variant="outline">{String(row.original.method ?? "linear")}</Badge>,
    },
    {
      accessorKey: "methodPeriod",
      header: "Period",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.methodPeriod ? String(row.original.methodPeriod) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => handleEdit(row.original)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  const isLoading = categoriesQuery.isLoading;
  const error = categoriesQuery.isError ? categoriesQuery.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={categories}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search categories…"}
        ariaLabel={"All categories"}
        emptyTitle={"No asset categories yet"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
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
            <span>{"Add category"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <AssetCategoryFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editCategory}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
