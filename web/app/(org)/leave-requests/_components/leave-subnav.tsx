"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const LEAVE_TABS = [
  { key: "leaveRequests", href: "/leave-requests" },
  { key: "leaveTypes", href: "/leave-requests/types" },
] as const;

export function LeaveSubNav() {
  const t = useTranslations("Leave");
  return <OrgSubNav label={t("subnavLabel")} tabs={LEAVE_TABS} />;
}
