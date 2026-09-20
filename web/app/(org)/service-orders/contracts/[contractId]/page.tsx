import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ServiceContractDetail } from "./_components/service-contract-detail-section";

export default async function ServiceContractDetailPage({
  params,
}: {
  params: Promise<{ contractId: string }>;
}) {
  const { contractId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/service-orders/contracts"}>{"Back to contracts"}</BackLink>
      <ServiceContractDetail orgId={id} contractId={contractId} />
    </div>
  );
}
