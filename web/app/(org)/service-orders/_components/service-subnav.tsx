"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SERVICE_TABS = [
  { key: "service", href: "/service-orders" },
  { key: "equipments", href: "/service-orders/equipment" },
  { key: "serviceContracts", href: "/service-orders/contracts" },
  { key: "maintenancePlans", href: "/service-orders/plans" },
] as const;

export function ServiceSubNav() {
  const t = useTranslations("Service");
  return <OrgSubNav label={t("title")} tabs={SERVICE_TABS} />;
}
