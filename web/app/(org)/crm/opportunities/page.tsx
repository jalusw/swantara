import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { OpportunitiesSection } from "../_components/opportunities-section";

export default async function OrgCrmOpportunitiesPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Opportunities"}
        description={"Qualified deals on the pipeline. Move them through stages to closing."}
      />
      <OpportunitiesSection orgId={id} />
    </div>
  );
}
