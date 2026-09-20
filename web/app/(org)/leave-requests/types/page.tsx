import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeaveTypesSection } from "./_components/leave-types-section";

export default async function OrgLeaveTypesPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Leave types"} description={"Description"} />
      <LeaveTypesSection orgId={id} />
    </div>
  );
}
