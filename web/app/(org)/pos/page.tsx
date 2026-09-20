import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosConfigsSection } from "./_components/pos-configs-section";

export default async function OrgPosConfigsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"POS Configurations"}
        description={"Configure your point of sale registers."}
      />
      <PosConfigsSection orgId={id} />
    </div>
  );
}
