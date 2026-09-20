"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const PRODUCT_TABS = [
  { key: "products", href: "/products" },
  { key: "productCategories", href: "/products/categories" },
  { key: "recipes", href: "/products/recipes" },
  { key: "price_books", href: "/products/price_books" },
  { key: "supplierCatalog", href: "/products/supplier-catalog" },
] as const;

export function ProductsSubNav() {
  return <OrgSubNav label={"Products"} tabs={PRODUCT_TABS} />;
}
