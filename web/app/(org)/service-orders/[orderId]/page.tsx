import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ServiceOrderDetail } from "./_components/service-order-detail-section";

export default async function ServiceOrderDetailPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/service-orders"}>{"Back to service orders"}</BackLink>
      <ServiceOrderDetail orgId={id} orderId={orderId} />
    </div>
  );
}
