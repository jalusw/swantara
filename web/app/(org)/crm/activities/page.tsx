import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ActivitiesSection } from "../_components/activities-section";

export default async function OrgCrmActivitiesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Activities"}
        description={"Calls, meetings, emails and tasks tied to a lead or contact."}
      />
      <ActivitiesSection orgId={id} />
    </div>
  );
}
