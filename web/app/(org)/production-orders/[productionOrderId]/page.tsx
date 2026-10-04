import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MoDetailSection } from "./_components/production-order-detail-section";

export default async function OrgProductionOrderDetailPage({
  params,
}: {
  params: Promise<{ productionOrderId: string }>;
}) {
  const t = await getTranslations("ProductionOrders");
  const { productionOrderId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/production-orders"}>{t("backToOrders")}</BackLink>
      <MoDetailSection orgId={id} productionOrderId={productionOrderId} />
    </div>
  );
}
