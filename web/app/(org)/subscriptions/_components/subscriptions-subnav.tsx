"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SUBSCRIPTIONS_TABS = [
  { key: "subscriptions", href: "/subscriptions" },
  { key: "subscriptionPlans", href: "/subscriptions/plans" },
] as const;

export function SubscriptionsSubNav() {
  return <OrgSubNav label={"Subscriptions"} tabs={SUBSCRIPTIONS_TABS} />;
}
