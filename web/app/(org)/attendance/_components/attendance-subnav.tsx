"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ATTENDANCE_TABS = [
  { key: "attendance", href: "/attendance" },
  { key: "timesheets", href: "/attendance/timesheets" },
] as const;

export function AttendanceSubNav() {
  return <OrgSubNav label={"Attendance"} tabs={ATTENDANCE_TABS} />;
}
