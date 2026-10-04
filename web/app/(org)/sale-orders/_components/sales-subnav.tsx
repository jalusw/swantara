"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SALES_TABS = [
  { key: "sales", href: "/sale-orders" },
  { key: "customers", href: "/sale-orders/customers" },
  { key: "rmas", href: "/sale-orders/returns" },
] as const;

export function SalesSubNav() {
  const t = useTranslations("Sales");
  return <OrgSubNav label={t("subnavLabel")} tabs={SALES_TABS} />;
}
