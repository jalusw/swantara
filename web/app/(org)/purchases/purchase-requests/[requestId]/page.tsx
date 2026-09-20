import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PurchaseRequestDetail } from "./_components/purchase-request-detail-section";

export default async function OrgPurchaseRequestDetailPage({
  params,
}: {
  params: Promise<{ requestId: string }>;
}) {
  const { requestId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/purchases/requisitions"}>{"Back to requisitions"}</BackLink>
      <PurchaseRequestDetail orgId={id} requestId={requestId} />
    </div>
  );
}
