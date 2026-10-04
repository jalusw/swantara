"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SUBSCRIPTIONS_TABS = [
  { key: "subscriptions", href: "/subscriptions" },
  { key: "subscriptionPlans", href: "/subscriptions/plans" },
] as const;

export function SubscriptionsSubNav() {
  const t = useTranslations("Subscriptions");
  return <OrgSubNav label={t("title")} tabs={SUBSCRIPTIONS_TABS} />;
}
