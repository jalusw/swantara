import { PageHeader } from "@/components/page-header";

import { ApprovalRequestsSection } from "./_components/approval-requests-section";

export default async function OrgApprovalRequestsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Approval requests"}
        description={"Review and decide on pending approval requests."}
      />
      <ApprovalRequestsSection />
    </div>
  );
}
