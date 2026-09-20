import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { InboundCostDetail } from "./_components/inbound-cost-detail-section";

export default async function InboundCostDetailPage({
  params,
}: {
  params: Promise<{ inboundCostId: string }>;
}) {
  const { inboundCostId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/purchases/inbound-costs"}>{"Back to landed costs"}</BackLink>
      <InboundCostDetail orgId={id} inboundCostId={inboundCostId} />
    </div>
  );
}
