import type { Metadata } from "next";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DashboardTabs } from "./_components/dashboard-tabs";

export const metadata: Metadata = {
  robots: {
    index: false,
    follow: false,
  },
};

export default async function OrgDashboardPage() {
  const id = String(await requireActiveOrgId());

  return <DashboardTabs orgId={id} />;
}
