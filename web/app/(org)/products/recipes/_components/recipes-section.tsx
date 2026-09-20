"use client";

import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, Recipe } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import type { StubBom, StubItem } from "../../_components/products-data";
import { BomFormDialog } from "./recipe-form-dialog";
import { BomsTable } from "./recipes-table";

function toBomRow(recipe: Recipe): StubBom {
  return {
    id: String(recipe.id),
    itemId: String(recipe.itemId),
    code: recipe.code,
    qty: recipe.qty,
    unitId: recipe.unitId != null ? String(recipe.unitId) : null,
    type: recipe.type,
    version: recipe.version,
    active: recipe.active,
    lines: [],
  };
}

function toProductRow(item: Item): StubItem {
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

export function BomsSection({ orgId }: { orgId: string }) {
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<StubBom | null>(null);

  const bomsQuery = useOrgListQuery<{ recipes: Recipe[] }, Record<string, never>>(
    "recipes",
    (organizationId) => getSwantaraService().recipes.list(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const recipes = (bomsQuery.data?.recipes ?? []).map(toBomRow);
  const templates = (productsQuery.data?.products ?? []).map(toProductRow);

  function handleSave() {
    setCreating(false);
    setEditing(null);
    void bomsQuery.refetch();
  }

  const isLoading = bomsQuery.isLoading || productsQuery.isLoading;
  const error = bomsQuery.isError
    ? bomsQuery.error
    : productsQuery.isError
      ? productsQuery.error
      : null;

  return (
    <div className="flex flex-col gap-4">
      <BomsTable
        recipes={recipes}
        templates={templates}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => {
                    void bomsQuery.refetch();
                    void productsQuery.refetch();
                  },
                }
              : undefined
        }
        renderActions={(recipe) => (
          <RowActions
            editLabel={"Edit"}
            deleteLabel={"Delete"}
            confirmTitle={"Delete this bill of materials?"}
            confirmDescription={"The bill of materials will be removed."}
            confirmLabel="OK"
            onEdit={() => setEditing(recipe)}
            onDelete={() => {
              void getSwantaraService()
                .recipes.delete(Number(orgId), Number(recipe.id))
                .then(() => void bomsQuery.refetch())
                .catch(() => toast.error("Could not disable the organization."));
            }}
          />
        )}
      />
      <div className="flex justify-end">
        <Button size="sm" onClick={() => setCreating(true)}>
          <Plus />
          <span>{"Add BoM"}</span>
        </Button>
      </div>
      {creating ? (
        <BomFormDialog
          open={creating}
          onOpenChange={setCreating}
          orgId={orgId}
          templates={templates}
          nextVersion={1}
          onSave={handleSave}
        />
      ) : null}
      {editing ? (
        <BomFormDialog
          open
          onOpenChange={() => setEditing(null)}
          orgId={orgId}
          templates={templates}
          initial={editing}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
