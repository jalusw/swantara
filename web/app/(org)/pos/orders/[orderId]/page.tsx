import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosOrderDetail } from "./_components/pos-order-detail-section";

export default async function OrgPosOrderDetailPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const t = await getTranslations("Pos");
  const { orderId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/pos/orders"}>{t("backToOrders")}</BackLink>
      <PosOrderDetail orgId={id} orderId={orderId} />
    </div>
  );
}
