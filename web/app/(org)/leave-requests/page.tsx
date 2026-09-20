import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeaveRequestsSection } from "./_components/leave-requests-section";

export default async function OrgLeaveRequestsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Leave requests"}
        description={"Manage employee leave requests and approvals."}
      />
      <LeaveRequestsSection orgId={id} />
    </div>
  );
}
