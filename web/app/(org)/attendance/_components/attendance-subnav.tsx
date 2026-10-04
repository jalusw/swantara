"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ATTENDANCE_TABS = [
  { key: "attendance", href: "/attendance" },
  { key: "timesheets", href: "/attendance/timesheets" },
] as const;

export function AttendanceSubNav() {
  const t = useTranslations("Attendance");
  return <OrgSubNav label={t("subnavLabel")} tabs={ATTENDANCE_TABS} />;
}
