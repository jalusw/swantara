"use client";

import { useTranslations } from "next-intl";
import { ActiveBadge } from "@/components/active-badge";
import { RecordLayout } from "@/components/record-layout";
import { useItemVariantsQuery, useProductQuery } from "@/lib/hooks/use-item-query";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Recipe } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

import { ProductBoms } from "./product-boms";
import { ProductOverview } from "./product-overview";
import { ItemVariants } from "./product-variants";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ProductDetail({ orgId, itemId }: { orgId: string; itemId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const tCommon = useTranslations("Common");
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
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!item) {
    return <p className="text-sm text-muted-foreground">{t("itemNotFound")}</p>;
  }

  const breadcrumbItems = [{ label: t("allProducts"), href: "/products" }, { label: item.name }];
  const tabs = [
    {
      id: "overview",
      label: t("overviewTab"),
      content: <ProductOverview item={item} />,
    },
    {
      id: "variants",
      label: t("variantsTab"),
      content: <ItemVariants variants={variants} />,
    },
    {
      id: "recipes",
      label: t("recipesTitle"),
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
      status={
        <ActiveBadge active={item.active}>{item.active ? t("active") : t("inactive")}</ActiveBadge>
      }
      tabs={tabs}
    />
  );
}
