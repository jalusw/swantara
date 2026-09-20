import { PageHeader } from "@/components/page-header";

import { MembersSection } from "./_components/members-section";

export default async function OrgMembersPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Members"}
        description={"Invite people to your organization and control their access."}
      />
      <MembersSection />
    </div>
  );
}
