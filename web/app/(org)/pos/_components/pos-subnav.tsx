"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const POS_TABS = [
  { key: "pos", href: "/pos" },
  { key: "posSessions", href: "/pos/sessions" },
  { key: "posOrders", href: "/pos/orders" },
  { key: "giftCards", href: "/pos/gift-cards" },
] as const;

export function PosSubNav() {
  const t = useTranslations("Pos");
  return <OrgSubNav label={t("title")} tabs={POS_TABS} />;
}
