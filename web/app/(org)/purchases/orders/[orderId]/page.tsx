import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PurchaseOrderDetail } from "./_components/purchase-order-detail-section";

export default async function OrgPurchaseOrderDetailPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/purchases"}>{"Back to purchase orders"}</BackLink>
      <PurchaseOrderDetail orgId={id} orderId={orderId} />
    </div>
  );
}
