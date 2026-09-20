import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InboundCostsSection } from "./_components/inbound-costs-section";

export default async function InboundCostsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Landed costs"}
        description={"Apply additional costs to incoming shipments."}
      />
      <InboundCostsSection orgId={id} />
    </div>
  );
}
