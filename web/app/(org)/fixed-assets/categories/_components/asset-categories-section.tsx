"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("FixedAssets");
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
      header: () => t("colName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "method",
      header: () => t("colDepreciationMethod"),
      cell: ({ row }) => (
        <Badge variant="outline">
          {(t as unknown as (k: string) => string)(
            `categoryMethod_${row.original.method ?? "linear"}`,
          )}
        </Badge>
      ),
    },
    {
      accessorKey: "methodPeriod",
      header: () => t("colPeriod"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.methodPeriod
            ? (t as unknown as (k: string) => string)(`methodPeriod_${row.original.methodPeriod}`)
            : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("actionEdit")}
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
        searchPlaceholder={t("searchCategories")}
        ariaLabel={t("allCategories")}
        emptyTitle={t("emptyCategories")}
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
            <span>{t("addCategory")}</span>
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
