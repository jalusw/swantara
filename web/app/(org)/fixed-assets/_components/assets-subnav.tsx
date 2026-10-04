"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ASSETS_TABS = [
  { key: "fixedAssets", href: "/fixed-assets" },
  { key: "assetCategories", href: "/fixed-assets/categories" },
] as const;

export function AssetsSubNav() {
  const t = useTranslations("FixedAssets");
  return <OrgSubNav label={t("layoutTitle")} tabs={ASSETS_TABS} />;
}
