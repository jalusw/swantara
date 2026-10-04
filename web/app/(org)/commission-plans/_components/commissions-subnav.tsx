"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const COMMISSIONS_TABS = [
  { key: "commissions", href: "/commission-plans" },
  { key: "commissionEntries", href: "/commission-plans/entries" },
] as const;

export function CommissionsSubNav() {
  const t = useTranslations("Commissions");
  return <OrgSubNav label={t("title")} tabs={COMMISSIONS_TABS} />;
}
