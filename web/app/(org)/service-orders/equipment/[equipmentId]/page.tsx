import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { EquipmentDetail } from "./_components/equipment-detail-section";

export default async function EquipmentDetailPage({
  params,
}: {
  params: Promise<{ equipmentId: string }>;
}) {
  const t = await getTranslations("Service");
  const { equipmentId } = await params;
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/service-orders/equipment"}>{t("backToEquipment")}</BackLink>
      <EquipmentDetail orgId={id} equipmentId={equipmentId} />
    </div>
  );
}
