"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const PRODUCT_TABS = [
  { key: "products", href: "/products" },
  { key: "productCategories", href: "/products/categories" },
  { key: "recipes", href: "/products/recipes" },
  { key: "price_books", href: "/products/price_books" },
  { key: "supplierCatalog", href: "/products/supplier-catalog" },
] as const;

export function ProductsSubNav() {
  const tNav = useTranslations("Nav");
  return <OrgSubNav label={tNav("products")} tabs={PRODUCT_TABS} />;
}
