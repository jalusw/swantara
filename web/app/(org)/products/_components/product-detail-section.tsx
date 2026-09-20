"use client";

import { ActiveBadge } from "@/components/active-badge";
import { RecordLayout } from "@/components/record-layout";
import { useItemVariantsQuery, useProductQuery } from "@/lib/hooks/use-item-query";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Recipe } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

import { ProductBoms } from "./product-boms";
import { ProductOverview } from "./product-overview";
import { ItemVariants } from "./product-variants";

export function ProductDetail({ orgId, itemId }: { orgId: string; itemId: string }) {
  const productQuery = useProductQuery(orgId, itemId);
  const variantsQuery = useItemVariantsQuery(orgId, itemId);

  const bomsQuery = useOrgListQuery<{ recipes: Recipe[] }, Record<string, never>, Recipe[]>(
    "recipes",
    (organizationId) => getSwantaraService().recipes.list(organizationId),
    {},
    { select: (data) => data.recipes.filter((recipe) => String(recipe.itemId) === itemId) },
  );

  const item = productQuery.data;
  const variants = variantsQuery.data ?? [];
  const recipes = bomsQuery.data ?? [];

  if (productQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!item) {
    return <p className="text-sm text-muted-foreground">{"Item not found."}</p>;
  }

  const breadcrumbItems = [{ label: "All products", href: "/products" }, { label: item.name }];
  const tabs = [
    {
      id: "overview",
      label: "Overview",
      content: <ProductOverview item={item} />,
    },
    {
      id: "variants",
      label: "Variants",
      content: <ItemVariants variants={variants} />,
    },
    {
      id: "recipes",
      label: "Bills of materials",
      content: (
        <ProductBoms
          recipes={recipes}
          status={
            bomsQuery.isLoading
              ? { type: "loading" }
              : bomsQuery.isError
                ? {
                    type: "error",
                    message: bomsQuery.error.message,
                    onRetry: () => void bomsQuery.refetch(),
                  }
                : undefined
          }
        />
      ),
    },
  ];

  return (
    <RecordLayout
      breadcrumbItems={breadcrumbItems}
      title={item.name}
      status={<ActiveBadge active={item.active}>{item.active ? "Active" : "Inactive"}</ActiveBadge>}
      tabs={tabs}
    />
  );
}
