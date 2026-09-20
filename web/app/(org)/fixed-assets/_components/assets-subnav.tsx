"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ASSETS_TABS = [
  { key: "fixedAssets", href: "/fixed-assets" },
  { key: "assetCategories", href: "/fixed-assets/categories" },
] as const;

export function AssetsSubNav() {
  return <OrgSubNav label={"Assets"} tabs={ASSETS_TABS} />;
}
