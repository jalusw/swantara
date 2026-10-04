import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { WarehouseDetail } from "./_components/warehouse-detail-section";

export default async function WarehouseDetailPage({
  params,
}: {
  params: Promise<{ warehouseId: string }>;
}) {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Stock",
  );
  const { warehouseId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/stock/warehouses"}>{t("backToWarehouses")}</BackLink>
      <WarehouseDetail orgId={id} warehouseId={warehouseId} />
    </div>
  );
}
