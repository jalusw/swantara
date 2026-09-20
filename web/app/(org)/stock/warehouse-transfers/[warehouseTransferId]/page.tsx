import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { WarehouseTransferDetail } from "./_components/warehouse-transfer-detail-section";

export default async function WarehouseTransferDetailPage({
  params,
}: {
  params: Promise<{ warehouseTransferId: string }>;
}) {
  const { warehouseTransferId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/stock/warehouse-transfers"}>{"Back to transfers"}</BackLink>
      <WarehouseTransferDetail orgId={id} warehouseTransferId={warehouseTransferId} />
    </div>
  );
}
