"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, ItemCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { activeColumn, nameColumn } from "@/lib/utils/table-columns";
import { ProductFormDialog } from "./product-form-dialog";
import type { StubItemCategory } from "./products-data";

function toProductRow(item: Item): {
  id: string;
  name: string;
  categoryId: string | null;
  type: Item["type"];
  unitId: string | null;
  purchaseUnitId: string | null;
  listPrice: number;
  standardCost: number;
  isPurchasable: boolean;
  isSellable: boolean;
  isManufactured: boolean;
  tracking: Item["tracking"];
  active: boolean;
} {
  return {
    id: String(item.id),
    name: item.name,
    categoryId: item.categoryId != null ? String(item.categoryId) : null,
    type: item.type,
    unitId: item.unitId != null ? String(item.unitId) : null,
    purchaseUnitId: item.purchaseUnitId != null ? String(item.purchaseUnitId) : null,
    listPrice: item.listPrice,
    standardCost: item.standardCost,
    isPurchasable: item.isPurchasable,
    isSellable: item.isSellable,
    isManufactured: item.isManufactured,
    tracking: item.tracking,
    active: item.active,
  };
}

function toCategoryRow(category: ItemCategory): StubItemCategory {
  return {
    id: String(category.id),
    name: category.name,
    parentId: category.parentId != null ? String(category.parentId) : null,
    incomeAccountId: category.incomeAccountId != null ? String(category.incomeAccountId) : null,
    expenseAccountId: category.expenseAccountId != null ? String(category.expenseAccountId) : null,
    stockCostAccountId:
      category.stockCostAccountId != null ? String(category.stockCostAccountId) : null,
    stockInputAccountId:
      category.stockInputAccountId != null ? String(category.stockInputAccountId) : null,
    stockOutputAccountId:
      category.stockOutputAccountId != null ? String(category.stockOutputAccountId) : null,
    cogsAccountId: category.cogsAccountId != null ? String(category.cogsAccountId) : null,
    costMethod: category.costMethod ?? null,
    valuation: category.valuation ?? null,
  };
}

function categoryName(categories: StubItemCategory[], id: string | null): string | undefined {
  return categories.find((category) => category.id === id)?.name;
}

export function ProductsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const categoriesQuery = useOrgListQuery<{ categories: ItemCategory[] }, Record<string, never>>(
    "productCategories",
    (organizationId) => getSwantaraService().productCategories.list(organizationId),
  );

  const templates = (productsQuery.data?.products ?? []).map(toProductRow);
  const categories = (categoriesQuery.data?.categories ?? []).map(toCategoryRow);

  function handleSave(itemId: string) {
    setDialogOpen(false);
    void productsQuery.refetch();
    router.push(`/products/${itemId}`);
  }

  const columns: ColumnDef<ReturnType<typeof toProductRow>>[] = [
    nameColumn<ReturnType<typeof toProductRow>>({
      basePath: "products",
      header: "Name",
    }),
    {
      accessorKey: "categoryId",
      header: "Category",
      cell: ({ row }) => (
        <Badge variant="secondary">
          {categoryName(categories, row.original.categoryId) ?? "—"}
        </Badge>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{humanizeKey(String(row.original.type))}</span>
      ),
    },
    {
      accessorKey: "listPrice",
      header: "Sales price",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.listPrice)}
        </span>
      ),
    },
    activeColumn<ReturnType<typeof toProductRow>>({
      header: "Status",
      activeLabel: "Active",
      inactiveLabel: "Inactive",
    }),
  ];

  const isLoading = productsQuery.isLoading || categoriesQuery.isLoading;
  const error = productsQuery.isError
    ? productsQuery.error
    : categoriesQuery.isError
      ? categoriesQuery.error
      : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={templates}
        getRowId={(row) => row.id}
        searchKeys={["name"]}
        statusKey="active"
        statusOptions={[
          { value: "true", label: "Active" },
          { value: "false", label: "Inactive" },
        ]}
        searchPlaceholder={"Search products…"}
        filterLabel={"Filter by status"}
        allLabel={"All statuses"}
        ariaLabel={"All products"}
        emptyTitle={"No products"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => {
                    void productsQuery.refetch();
                    void categoriesQuery.refetch();
                  },
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Add item"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ProductFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          categories={categories.map(({ id, name }) => ({ id, name }))}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
