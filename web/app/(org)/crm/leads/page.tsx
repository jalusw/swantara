import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeadsSection } from "../_components/leads-section";

export default async function OrgCrmLeadsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Leads"}
        description={"Unqualified contacts. Promote a lead to create an opportunity."}
      />
      <LeadsSection orgId={id} />
    </div>
  );
}
