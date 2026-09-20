"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const LEAVE_TABS = [
  { key: "leaveRequests", href: "/leave-requests" },
  { key: "leaveTypes", href: "/leave-requests/types" },
] as const;

export function LeaveSubNav() {
  return <OrgSubNav label={"Leave"} tabs={LEAVE_TABS} />;
}
