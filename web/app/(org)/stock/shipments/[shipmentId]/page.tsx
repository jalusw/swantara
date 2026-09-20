import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ShipmentDetail } from "./_components/shipment-detail-section";

export default async function ShipmentDetailPage({
  params,
}: {
  params: Promise<{ shipmentId: string }>;
}) {
  const { shipmentId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/stock/shipments"}>{"Back to shipments"}</BackLink>
      <ShipmentDetail orgId={id} shipmentId={shipmentId} />
    </div>
  );
}
